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

func (s *store) AcceptOrder(ctx context.Context, order *pb.Order) error {
	log.Printf("Accepting order %s for customer %s", order.ID, order.CustomerID)

	if order == nil {
		return fmt.Errorf("order is nil")
	}
	if order.ID == "" {
		return fmt.Errorf("order ID is required")
	}
	if order.CustomerID == "" {
		return fmt.Errorf("customer ID is required")
	}
	if order.Status == "" {
		return fmt.Errorf("order status is required")
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("order must contain at least one item")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert into orders table
	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (id, customer_id, status, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, order.ID, order.CustomerID, order.Status)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert all order items
	for _, item := range order.Items {
		if item == nil {
			return fmt.Errorf("order item is nil")
		}
		if item.ID == "" {
			return fmt.Errorf("item ID is required")
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item quantity must be positive (item %s)", item.ID)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO order_items (order_id, item_id, quantity)
			VALUES (?, ?, ?)
		`, order.ID, item.ID, item.Quantity)
		if err != nil {
			return fmt.Errorf("failed to insert order item (order=%s item=%s): %w",
				order.ID, item.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

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

	// Delete order items first (to satisfy FK constraints if you add them later)
	res, err := tx.ExecContext(ctx, `
		DELETE FROM order_items
		WHERE order_id = ?
	`, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order items: %w", err)
	}

	// Optional: check whether any items were actually deleted
	_, _ = res.RowsAffected() // ignore or log if you want

	// Delete the order itself
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

func (s *store) Close() error {
	return s.db.Close()
}
