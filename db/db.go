package db

import (
	"log"
	"os"
	"github.com/go-pg/pg/v9"

	controller "github.com/amirsorouri00/arvanwallet/controller"
)


// Connecting to DB
func ConnectDB() *pg.DB {
	opts := &pg.Options {
		User: "go_db",
		Password: "123123",
		Addr: "localhost:5432",
		Database: "go_db",
	}

	var db *pg.DB = pg.Connect(opts)
	if db == nil {
		log.Printf("Failed to connect")
		os.Exit(100)
	}

	log.Printf("Connected to DB")
	controller.CreateTransactionTable(db)
	controller.CreateUserTable(db)

	// Pass DB Connection to the controller
	controller.InitiateDB(db)
	return db
}


