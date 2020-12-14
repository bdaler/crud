package main

import (
	"context"
	"github.com/bdaler/crud/cmd/app"
	"github.com/bdaler/crud/pkg/customers"
	"github.com/bdaler/crud/pkg/security"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/dig"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	HOST = "0.0.0.0"
	PORT = "9999"
)

func main() {
	dsn := "postgres://app:pass@localhost:5432/db"
	if err := execute(HOST, PORT, dsn); err != nil {
		os.Exit(1)
	}
}

func execute(server, port, dsn string) (err error) {
	deps := []interface{}{
		app.NewServer,
		mux.NewRouter,
		customers.NewService,
		security.NewService,
		func() (*pgxpool.Pool, error) {
			connCtx, _ := context.WithTimeout(context.Background(), time.Second*5)
			return pgxpool.Connect(connCtx, dsn)
		},
		func(serverHandler *app.Server) *http.Server {
			return &http.Server{
				Addr:    net.JoinHostPort(server, port),
				Handler: serverHandler,
			}
		},
	}

	container := dig.New()
	for _, dep := range deps {
		err = container.Provide(dep)
		if err != nil {
			return err
		}
	}

	err = container.Invoke(func(app *app.Server) { app.Init() })
	if err != nil {
		return err
	}

	return container.Invoke(func(s *http.Server) error { return s.ListenAndServe() })
}
