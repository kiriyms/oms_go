package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	pb "github.com/kiriyms/oms_go-common/api"
	_ "github.com/mattn/go-sqlite3"
)

type Store interface {
	AcceptOrder(context.Context, *pb.Order) error
	FinishOrder(context.Context, string) error
	GetOrder(context.Context, string) (*pb.Order, error)
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

func (s *store) AcceptOrder(ctx context.Context, o *pb.Order) error {
	log.Printf("Accepting order %+v", o)

	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("BEGIN FAILED: %v", err)
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO orders (id, customer_id, status)
		VALUES (?, ?, ?)
	`, o.ID, o.CustomerID, o.Status)
	if err != nil {
		log.Printf("ORDER INSERT FAILED: %v", err)
		return fmt.Errorf("failed to insert order: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO order_items (order_id, item_id, quantity)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		log.Printf("PREPARE FAILED: %v", err)
		return fmt.Errorf("failed to prepare order_items stmt: %w", err)
	}
	defer stmt.Close()

	for _, item := range o.Items {
		if item.Quantity <= 0 {
			log.Printf("INVALID QUANTITY: %d", item.Quantity)
			return fmt.Errorf("invalid quantity %d for item %s", item.Quantity, item.ID)
		}

		_, err := stmt.Exec(o.ID, item.ID, item.Quantity)
		if err != nil {
			log.Printf("ITEM INSERT FAILED (item=%s): %v", item.ID, err)
			return fmt.Errorf("failed to insert order item %s: %w", item.ID, err)
		}
	}

	log.Printf("COMMITTING ORDER %s", o.ID)
	if err := tx.Commit(); err != nil {
		log.Printf("COMMIT FAILED: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("ORDER %s COMMITTED OK", o.ID)
	return nil
}

func (s *store) FinishOrder(ctx context.Context, orderID string) error {
	log.Printf("Finishing order %s", orderID)

	if orderID == "" {
		return fmt.Errorf("order ID is required")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		DELETE FROM order_items
		WHERE order_id = ?
	`, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order items: %w", err)
	}

	_, _ = res.RowsAffected()

	res, err = tx.ExecContext(ctx, `
		DELETE FROM orders
		WHERE id = ?
	`, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("order %s not found", orderID)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *store) GetOrder(ctx context.Context, orderID string) (*pb.Order, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	var (
		id         string
		customerID string
		status     string
	)

	err := s.db.QueryRowContext(ctx, `
		SELECT id, customer_id, status
		FROM orders
		WHERE id = ?
	`, orderID).Scan(&id, &customerID, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order %s not found", orderID)
		}
		return nil, fmt.Errorf("failed to query order: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT item_id, quantity
		FROM order_items
		WHERE order_id = ?
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []*pb.Item
	for rows.Next() {
		var (
			itemID   string
			quantity int32
		)

		if err := rows.Scan(&itemID, &quantity); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}

		items = append(items, &pb.Item{
			ID:       itemID,
			Quantity: quantity,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	order := &pb.Order{
		ID:         id,
		CustomerID: customerID,
		Status:     status,
		Items:      items,
	}

	return order, nil
}

func (s *store) Close() error {
	return s.db.Close()
}
