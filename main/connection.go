package main

import (
	"log"

	"github.com/go-pg/pg/v10"
)

func connectToDatabase() *pg.DB {
	opt, err := pg.ParseURL("postgres://prabik98:incorrect@localhost:5432/PostgreSQL")
	if err != nil {
		log.Fatal("Failed to parse database URL:", err)
	}

	db := pg.Connect(opt)
	if db == nil {
		log.Fatal("Failed to connect to the database")
	}
	return db
}
