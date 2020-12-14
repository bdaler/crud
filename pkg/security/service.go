package security

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Auth(login, password string) bool {
	sqlStatement := `select login, password from managers where login=$1 and password=$2`
	err := s.db.QueryRow(context.Background(), sqlStatement, login, password).Scan(&login, &password)
	if err != nil {
		log.Print(err)
		return false
	}
	return true
}
