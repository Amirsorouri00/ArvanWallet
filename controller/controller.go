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
	Id        		string    		`json:"id"`	// #CellPhone
	Cash	  		float64			`json:"cash"`
	Transactions	[]*Transaction	
	CreatedAt 		time.Time 		
	UpdatedAt 		time.Time 				
}

type Transaction struct {
	Id        string    `json:"id"`
	Amount	  float64	`json:"amount"`
	Type 	  bool		`json:"type"` // [false = Increase, true = Decrease]
	Gift	  bool		`json:"gift"` // [false = No, true = Yes]
	GiftId	  string	`json:",use_zero"` // zero/GiftID
	UserId	  string			
	CreatedAt time.Time 
	UpdatedAt time.Time 
}


// Create User Table OK
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


// Create Transaction Table OK
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


// Get All Users OK
func GetAllUsers(c *gin.Context) {
	var users []User
	err := dbConnect.Model(&users).Relation("Transactions").Select()

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


// Add User OK
func AddUser(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	id := user.Id

	insertError := dbConnect.Insert(&User{
		Id: id,
		Cash: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
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


// Get Single User(internal function) OK
func GetSingleUser(user *User) int {
	err := dbConnect.Select(user)
	if err != nil {
		log.Printf("Error while getting a single todo, Reason: %v\n", err)
		log.Printf("user id = %v\n", user.Id)
		return 0
	}
	return 1
}


// Get All Transactions OK
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


// Struct related to the Add Transaction API
// https://mholt.github.io/json-to-go/
type AddTransactionType struct {
	Amount float64  `json:"amount"`
	Type   bool  	`json:"type"`
	User   string   `json:"user"`
}

// Add Transaction API OK
func AddTransaction(c *gin.Context) {
	var transaction AddTransactionType
	c.BindJSON(&transaction)
	amount := transaction.Amount
	transType := transaction.Type
	userId := transaction.User

	user := &User{Id: userId}
	err := GetSingleUser(user)
	log.Printf("user id = %v and cash = %v\n", user.Id, user.Cash)
	if err == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "User in the request is not a valid user",
		})
		return
	} else {
		updateCash := 0.0
		if transType /*== "true"*/ {
			// It means the transaction amount must
			// decrease from the user cash.
			log.Printf("Hear transType == true\n")
			if user.Cash < amount {
				log.Printf("Invalid amount for the transaction. User cash is less than the amount of the transaction to decrease")
			}
			updateCash = user.Cash - amount
		} else {
			log.Printf("Hear transType == false\n")
			updateCash = user.Cash + amount
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
			Id: guuid.New().String(),
			Amount: amount,
			Type: transType,
			Gift: false,
			UserId: user.Id,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
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