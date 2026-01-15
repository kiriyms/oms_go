package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	pb "github.com/kiriyms/oms_go-common/api"
	_ "github.com/mattn/go-sqlite3"
)

type OrderStore interface {
	Create(context.Context, *pb.Order) error
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

func (s *store) Close() error {
	err := s.db.Close()
	return err
}
