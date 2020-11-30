package customers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"
)

var ErrNotFound = errors.New("customer not found")
var ErrInternal = errors.New("internal error")

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type Customer struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Phone   string    `json:"phone"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
}

func (s *Service) ByID(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}

	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, phone, active, created FROM customers WHERE id = $1
	`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Println(err)
		return nil, ErrInternal
	}
	return item, nil
}

func (s *Service) All(ctx context.Context) (customers []*Customer, err error) {
	sqlStatement := `SELECT * FROM customers`
	rows, err := s.db.QueryContext(ctx, sqlStatement)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &Customer{}
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Phone,
			&item.Active,
			&item.Created,
		)

		if err != nil {
			log.Println(err)
		}

		customers = append(customers, item)
	}

	return customers, nil
}

func (s *Service) AllActive(ctx context.Context) (customers []*Customer, err error) {
	sqlStatement := `SELECT * FROM customers WHERE active = TRUE`
	rows, err := s.db.QueryContext(ctx, sqlStatement)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &Customer{}
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Phone,
			&item.Active,
			&item.Created,
		)
		if err != nil {
			log.Println(err)
		}
		customers = append(customers, item)
	}

	return customers, nil
}
func (s *Service) ChangeActive(ctx context.Context, id int64, active bool) (*Customer, error) {
	item := &Customer{}

	sqlStatement := `UPDATE customers SET active=$2 WHERE id=$1 RETURNING *`
	err := s.db.QueryRowContext(ctx, sqlStatement, id, active).Scan(
		&item.ID,
		&item.Name,
		&item.Phone,
		&item.Active,
		&item.Created)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil
}

func (s *Service) Delete(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}

	sqlStatement := `DELETE FROM customers WHERE id=$1 RETURNING *`
	err := s.db.QueryRowContext(ctx, sqlStatement, id).Scan(
		&item.ID,
		&item.Name,
		&item.Phone,
		&item.Active,
		&item.Created)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil
}

func (s *Service) Save(ctx context.Context, customer *Customer) (c *Customer, err error) {
	item := &Customer{}

	if customer.ID == 0 {
		sqlStatement := `INSERT INTO customers(name, phone) VALUES ($1, $2) RETURNING *`
		err = s.db.QueryRowContext(ctx, sqlStatement, customer.Name, customer.Phone).Scan(
			&item.ID,
			&item.Name,
			&item.Phone,
			&item.Active,
			&item.Created)
	} else {
		sqlStatement := `UPDATE customers SET name=$1, phone=$2 WHERE id=$3 RETURNING *`
		err = s.db.QueryRowContext(ctx, sqlStatement, customer.Name, customer.Phone, customer.ID).Scan(
			&item.ID,
			&item.Name,
			&item.Phone,
			&item.Active,
			&item.Created)
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil
}
