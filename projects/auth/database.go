package main

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)


func connect(url string) (*sql.DB, error) {
	db,err := sql.Open("pgx",url)

	if err != nil {
		return nil,err
	}

	err = db.Ping()
	if err != nil {
		return nil,err
	}

	return db,nil
}
