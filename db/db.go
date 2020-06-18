package db

import (
	"log"
	"os"
	"github.com/gin-gonic/gin"
	orm "github.com/go-pg/pg/v9/orm"
	"github.com/go-pg/pg/v9"
	guuid "github.com/google/uuid"

	//controller "./controller"
)


// Connecting to DB
func ConnectDB() *pg.DB {
	opts := &pg.options {
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
	CreateUserTable(db)
	CreateTransactionTable(db)

	// Pass DB Connection to the controller
	controller.InitiateDB(db)
	return db
}


type User struct {
	ID        		string    		`json:"id"`	// #CellPhone
	Cash	  		float64			`json:"cash"`
	//Transactions	[]Transaction	`json:"fk:transaction_id`
	CreatedAt 		time.Time 		`json:"default:now()"`
	UpdatedAt 		time.Time 		`json:"default:now()"`		
}

type Transaction struct {
	ID        string    `json:"id"`
	Amount	  float64	`json:"amount"`
	Type 	  bool		`json:"type"` // [0 == Increase, 1 == Decrease]
	Gift	  bool		`json:"gift"` // [0 == No, 1 == Yes]
	GiftId	  string	`json:",use_zero"` // zero/GiftID
	UserId	  User		`json:"fk:user_id"`	
	CreatedAt time.Time `json:"default:now()"`
	UpdatedAt time.Time `json:"default:now()"`
}

// Create User Table
func CreateUserTable(db *pg.DB) error {
	opts := &orm.CreateTableOptions{
		IfNotExists: true,
	}
	createError := db.CreateTable(&User{}, opts)
	if createError != nil {
		log.Printf("Error while creating User table, Reason: %v\n", createError)
		return createError
	}
	log.Printf("User table created.")
	return nil
}

// Create Transaction Table
func CreateTransactionTable(db *pg.DB) error {
	opts := &orm.CreateTableOptions{
		IfNotExists: true,
	}
	createError := db.CreateTable(&Transaction{}, opts)
	if createError != nil {
		log.Printf("Error while creating Transaction table, Reason: %v\n", createError)
		return createError
	}
	log.Printf("Transaction table created.")
	return nil
}