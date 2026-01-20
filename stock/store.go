package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	pb "github.com/kiriyms/oms_go-common/api"
	_ "github.com/mattn/go-sqlite3"
)

type StockStore interface {
	AddStockItem(ctx context.Context, item *pb.StockItem) (*pb.StockItem, error)
	BookStockItem(ctx context.Context, itemID string, quantity int32) (*pb.ItemWithQuantity, error)
	ReleaseBookItem(ctx context.Context, itemID string, quantity int32) (*pb.ItemWithQuantity, error)
	RemoveStockItem(ctx context.Context, itemID string) (*pb.StockItem, error)
	VerifyStock(ctx context.Context, items []*pb.ItemWithQuantity) *pb.VerifyStockResponse
	GetStockItem(ctx context.Context, itemID string) (*pb.StockItem, error)
	FinalizeBooking(ctx context.Context, orderID string) error
	Close() error
}

type store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	s := &store{db: db}
	return s, nil
}

func (s *store) AddStockItem(ctx context.Context, item *pb.StockItem) (*pb.StockItem, error) {
	log.Printf("Adding stock item: %+v", item)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO stock_items (id, quantity, name, price_id, description, img_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(id)
		DO UPDATE SET
			quantity   = stock_items.quantity + excluded.quantity,
			name       = excluded.name,
			price_id   = excluded.price_id,
			description= excluded.description,
			img_path   = excluded.img_path,
			updated_at = CURRENT_TIMESTAMP
	`,
		item.ID,
		item.Quantity,
		item.Name,
		item.PriceID,
		item.Description,
		item.ImgPath,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to add stock item: %w", err)
	}

	return item, nil
}

func (s *store) BookStockItem(ctx context.Context, itemID string, quantity int32) (*pb.ItemWithQuantity, error) {
	log.Printf("Booking stock item %s qty %d", itemID, quantity)

	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var stockQty int32
	err = tx.QueryRowContext(ctx, `
		SELECT quantity
		FROM stock_items
		WHERE id = ?
	`, itemID).Scan(&stockQty)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("item %s not found", itemID)
		}
		return nil, err
	}

	var bookedQty int32
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(quantity), 0)
		FROM booked_items
		WHERE item_id = ?
		  AND expires_at > CURRENT_TIMESTAMP
	`, itemID).Scan(&bookedQty)
	if err != nil {
		return nil, err
	}

	available := stockQty - bookedQty
	if available < quantity {
		return nil, fmt.Errorf("insufficient stock: available=%d requested=%d", available, quantity)
	}

	expiresAt := time.Now().Add(15 * time.Minute)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO booked_items (booking_id, item_id, quantity, order_id, expires_at, created_at)
		VALUES (lower(hex(randomblob(16))), ?, ?, '', ?, CURRENT_TIMESTAMP)
	`,
		itemID,
		quantity,
		expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to book item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &pb.ItemWithQuantity{
		ID:       itemID,
		Quantity: quantity,
	}, nil
}

func (s *store) ReleaseBookItem(ctx context.Context, itemID string, quantity int32) (*pb.ItemWithQuantity, error) {
	log.Printf("Releasing booking for item %s qty %d", itemID, quantity)

	if quantity <= 0 {
		_, err := s.db.ExecContext(ctx, `
			DELETE FROM booked_items
			WHERE item_id = ?
		`, itemID)
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT booking_id, quantity
		FROM booked_items
		WHERE item_id = ?
		ORDER BY created_at ASC
	`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	remaining := quantity

	for rows.Next() && remaining > 0 {
		var bookingID string
		var q int32

		if err := rows.Scan(&bookingID, &q); err != nil {
			return nil, err
		}

		if q <= remaining {
			_, err = tx.ExecContext(ctx, `
				DELETE FROM booked_items WHERE booking_id = ?
			`, bookingID)
			if err != nil {
				return nil, err
			}
			remaining -= q
		} else {
			_, err = tx.ExecContext(ctx, `
				UPDATE booked_items
				SET quantity = quantity - ?
				WHERE booking_id = ?
			`, remaining, bookingID)
			if err != nil {
				return nil, err
			}
			remaining = 0
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &pb.ItemWithQuantity{
		ID:       itemID,
		Quantity: quantity,
	}, nil
}

func (s *store) RemoveStockItem(ctx context.Context, itemID string) (*pb.StockItem, error) {
	log.Printf("Removing stock item %s", itemID)

	item, err := s.GetStockItem(ctx, itemID)
	if err != nil {
		return nil, err
	}

	_, err = s.db.ExecContext(ctx, `
		DELETE FROM stock_items
		WHERE id = ?
	`, itemID)

	if err != nil {
		return nil, fmt.Errorf("failed to remove stock item: %w", err)
	}

	return item, nil
}

func (s *store) VerifyStock(ctx context.Context, items []*pb.ItemWithQuantity) *pb.VerifyStockResponse {
	resp := &pb.VerifyStockResponse{
		AllAvailable:          true,
		MissingOrInsufficient: []*pb.ItemWithQuantity{},
	}

	for _, item := range items {
		var stockQty int32
		err := s.db.QueryRowContext(ctx, `
			SELECT quantity
			FROM stock_items
			WHERE id = ?
		`, item.ID).Scan(&stockQty)

		if err != nil {
			resp.AllAvailable = false
			resp.MissingOrInsufficient = append(resp.MissingOrInsufficient, &pb.ItemWithQuantity{
				ID:       item.ID,
				Quantity: item.Quantity,
			})
			continue
		}

		var bookedQty int32
		err = s.db.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(quantity), 0)
			FROM booked_items
			WHERE item_id = ?
			  AND expires_at > CURRENT_TIMESTAMP
		`, item.ID).Scan(&bookedQty)

		if err != nil {
			resp.AllAvailable = false
			continue
		}

		available := stockQty - bookedQty
		if available < item.Quantity {
			resp.AllAvailable = false
			resp.MissingOrInsufficient = append(resp.MissingOrInsufficient, &pb.ItemWithQuantity{
				ID:       item.ID,
				Quantity: item.Quantity,
			})
		}
	}

	return resp
}

func (s *store) GetStockItem(ctx context.Context, itemID string) (*pb.StockItem, error) {
	var item pb.StockItem

	err := s.db.QueryRowContext(ctx, `
		SELECT id, quantity, name, price_id, description, img_path, created_at, updated_at
		FROM stock_items
		WHERE id = ?
	`, itemID).Scan(
		&item.ID,
		&item.Quantity,
		&item.Name,
		&item.PriceID,
		&item.Description,
		&item.ImgPath,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("item %s not found", itemID)
		}
		return nil, fmt.Errorf("failed to fetch stock item: %w", err)
	}

	return &item, nil
}

func (s *store) FinalizeBooking(ctx context.Context, orderID string) error {
	log.Printf("Finalizing booking for order %s", orderID)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT item_id, SUM(quantity) AS total_qty
		FROM booked_items
		WHERE order_id = ?
		  AND expires_at > CURRENT_TIMESTAMP
		GROUP BY item_id
	`, orderID)
	if err != nil {
		return fmt.Errorf("failed to load bookings: %w", err)
	}
	defer rows.Close()

	type itemAgg struct {
		itemID string
		qty    int32
	}

	var items []itemAgg

	for rows.Next() {
		var it itemAgg
		if err := rows.Scan(&it.itemID, &it.qty); err != nil {
			return err
		}
		items = append(items, it)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if len(items) == 0 {
		return fmt.Errorf("no active bookings found for order %s", orderID)
	}

	for _, it := range items {
		var stockQty int32
		err := tx.QueryRowContext(ctx, `
			SELECT quantity
			FROM stock_items
			WHERE id = ?
		`, it.itemID).Scan(&stockQty)
		if err != nil {
			return fmt.Errorf("stock item %s not found: %w", it.itemID, err)
		}

		if stockQty < it.qty {
			return fmt.Errorf("insufficient stock during finalize: item=%s have=%d need=%d",
				it.itemID, stockQty, it.qty)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE stock_items
			SET quantity = quantity - ?,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, it.qty, it.itemID)
		if err != nil {
			return fmt.Errorf("failed to deduct stock for %s: %w", it.itemID, err)
		}
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM booked_items
		WHERE order_id = ?
	`, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete bookings: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Booking finalized successfully for order %s", orderID)
	return nil
}

func (s *store) Close() error {
	return s.db.Close()
}
