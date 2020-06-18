package controller

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	orm "github.com/go-pg/pg/v9/orm"
	"github.com/go-pg/pg/v9"
	guuid "github.com/google/uuid"
)

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
	Type 	  bool		`json:"type"` // [false == Increase, true == Decrease]
	Gift	  bool		`json:"gift"` // [false == No, true == Yes]
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

// INITIALIZE DB CONNECTION (TO AVOID TOO MANY CONNECTION)
var dbConnect *pg.DB
func InitiateDB(db *pg.DB) {
	dbConnect = db
}

func GetAllUsers(c *gin.Context) {
	var users []User
	err := dbConnect.Model(&users).Select()

	if err != nil {
		log.Printf("Error while getting all users, Reason: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "All Users",
		"data": users,
	})
	return
}

func AddUser(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	id := user.ID

	insertError := dbConnect.Insert(&User{
		ID: id,
		Cash: 0,
	})
	if insertError != nil {
		log.Printf("Error while inserting new user into db, Reason: %v\n", insertError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "User created Successfully",
	})
	return
}

func GetSingleUser(user *User) int {
	err := dbConnect.Select(user)
	if err != nil {
		log.Printf("Error while getting a single todo, Reason: %v\n", err)
		log.Printf("user id = %v\n", user.ID)
		return 0
	}
	return 1
}

func GetAllTransactions(c *gin.Context) {
	var transactions []Transaction
	err := dbConnect.Model(&transactions).Select()

	if err != nil {
		log.Printf("Error while getting all transactions, Reason: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return 
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "All transactions",
		"data": transactions,
	})
	return
}

type transaction_form struct {
	transaction		Transaction		`json:"transaction"`
	userId			string			`json:"user"`
}

func AddTransaction(c *gin.Context) {
	var transaction transaction_form
	c.BindJSON(&transaction)
	amount := transaction.transaction.Amount
	transType := transaction.transaction.Type
	userId := transaction.userId

	// var user1 User
	// c.BindJSON(&user1)
	// userId := user1.ID
	// userId := c.Param("user")
	user := &User{ID: userId}
	log.Printf("user id0 = %v\n", userId)
	err := GetSingleUser(user)
	if err == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "User in the request is not a valid user",
		})
		return
	} else {
		updateCash := 0.0
		if transType {
			if user.Cash < amount {
				log.Printf("Invalid amount for the transaction. User cash is less than the amount of the transaction to decrease")
			}
			updateCash = user.Cash - amount
		} else {
			updateCash = user.Cash - amount
		}

		_, error1 := dbConnect.Model(&User{}).Set("cash = ?", updateCash).Where("id = ?", userId).Update()
		if error1 != nil {
			log.Printf("Error, Reason: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": 500,
				"message":  "Something went wrong while updating user cash.",
			})
			return
		}

		insertError := dbConnect.Insert(&Transaction{
			ID: guuid.New().String(),
			Amount: amount,
			Type: transType,
			Gift: false,
			UserId: *user,
		})
		if insertError != nil {
			log.Printf("Error while inserting new transaction into db, Reason: %v\n", insertError)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Something went wrong while adding transaction",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status":  http.StatusCreated,
			"message": "Transaction created Successfully",
		})
		return
	}

	
}