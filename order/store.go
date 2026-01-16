package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	pb "github.com/kiriyms/oms_go-common/api"
	_ "github.com/mattn/go-sqlite3"
)

type OrderStore interface {
	Create(context.Context, *pb.Order) error
	GetOrder(context.Context, string) (*pb.Order, error)
	GetUserOrders(context.Context, string) ([]*pb.Order, error)
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

func (s *store) Create(ctx context.Context, o *pb.Order) error {
	log.Printf("Creating order in the database: %+v", o)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO orders (id, customer_id, status)
		VALUES (?, ?, ?)
	`,
		o.ID,
		o.CustomerID,
		o.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO order_items (order_id, item_id, quantity)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare order_items stmt: %w", err)
	}
	defer stmt.Close()

	for _, item := range o.Items {
		if item.Quantity <= 0 {
			return fmt.Errorf("invalid quantity %d for item %s", item.Quantity, item.ID)
		}

		_, err := stmt.Exec(o.ID, item.ID, item.Quantity)
		if err != nil {
			return fmt.Errorf("failed to insert order item %s: %w", item.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *store) GetOrder(ctx context.Context, orderID string) (*pb.Order, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var o pb.Order

	err = tx.QueryRowContext(ctx, `
		SELECT id, customer_id, status
		FROM orders
		WHERE id = ?
	`, orderID).Scan(
		&o.ID,
		&o.CustomerID,
		&o.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order %s not found", orderID)
		}
		return nil, fmt.Errorf("failed to fetch order: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT item_id, quantity
		FROM order_items
		WHERE order_id = ?
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item pb.Item
		if err := rows.Scan(&item.ID, &item.Quantity); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		o.Items = append(o.Items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &o, nil
}

func (s *store) GetUserOrders(ctx context.Context, userID string) ([]*pb.Order, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT id, customer_id, status
		FROM orders
		WHERE customer_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user orders: %w", err)
	}
	defer rows.Close()

	orderMap := make(map[string]*pb.Order)
	var orderIDs []string

	for rows.Next() {
		var o pb.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.Status); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orderMap[o.ID] = &o
		orderIDs = append(orderIDs, o.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(orderIDs) == 0 {
		return []*pb.Order{}, nil
	}

	query, args := buildInQuery(`
		SELECT order_id, item_id, quantity
		FROM order_items
		WHERE order_id IN (%s)
	`, orderIDs)

	itemRows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order items: %w", err)
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var orderID string
		var item pb.Item

		if err := itemRows.Scan(&orderID, &item.ID, &item.Quantity); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}

		orderMap[orderID].Items = append(orderMap[orderID].Items, &item)
	}

	if err := itemRows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	orders := make([]*pb.Order, 0, len(orderMap))
	for _, o := range orderMap {
		orders = append(orders, o)
	}

	return orders, nil
}

func (s *store) Close() error {
	err := s.db.Close()
	return err
}

func buildInQuery(base string, ids []string) (string, []interface{}) {
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))

	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	return fmt.Sprintf(base, strings.Join(placeholders, ",")), args
}
