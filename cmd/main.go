package main

import (
	"github.com/bdaler/crud/cmd/server/app"
	"github.com/bdaler/crud/pkg/customers"
	"database/sql"
	_ "github.com/jackc/pgx/v4"
	"log"
	"net"
	"net/http"
	"os"
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
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := db.Close(); cerr != nil{
			if err == nil {
				err = cerr
				return
			}
			log.Println(err)
		}
	}()

	mux := http.NewServeMux()
	customerSvs := customers.NewService(db)
	serverHandler := app.NewServer(mux, customerSvs)
	serverHandler.Init()

	srv := &http.Server{
		Addr:    net.JoinHostPort(server, port),
		Handler: serverHandler,
	}
	return srv.ListenAndServe()
}