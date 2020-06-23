package controller

import (
	"log"
	"fmt"
	"time"
	"bytes"
	"net/http"
	"io/ioutil"
	"encoding/json"

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

// See Cash in Wallet OK
func SeeWalletCash(c *gin.Context) {
	var user User
	c.BindJSON(&user)

	err := dbConnect.Model(&user).Where("id = ?", user.Id).Select()
	if err != nil {
		log.Printf("SeeWalletCash: Error while getting user's data from db, Reason: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "User cash in wallet received",
		"cash": user.Cash,
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
			log.Printf("AddTransaction: Error while inserting new transaction into db, Reason: %v\n", insertError)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Something went wrong while adding transaction",
			})
			return
		}

		_, error1 := dbConnect.Model(&User{}).Set("cash = ?", updateCash).Where("id = ?", userId).Update()
		if error1 != nil {
			log.Printf("AddTransaction: Error, Reason: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": 500,
				"message":  "Something went wrong while updating user cash.",
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


type GiftChargeType struct {
	UserId   string `json:"user_id"`
	GiftCode string `json:"gift_code"`
}

type GetGiftType struct {
	Code string `json:"code"`
}

type PostReqReturnType struct {
	Status     int      `json:"status"`
	GiftAmount float64  `json:"gift_amount"`
}

// Gift Charge OK
func GiftCharge(c *gin.Context) {
	var req GiftChargeType
	c.BindJSON(&req)

	var user User
	err := dbConnect.Model(&user).Where("id = ?", req.UserId).Select()
	if err != nil {
		log.Printf("GiftCharge: Error while getting user, Reason: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	var trans Transaction
	count, err2 := dbConnect.Model(&trans).Where("gift_id = ?", req.GiftCode).Where("user_id = ?", req.UserId).Count()
	if err2 != nil {
		log.Printf("GiftCharge: Error while getting user, Reason: %v\n", err2)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	if count > 0 {
		log.Printf("GiftCharge: This user has charged his wallet with this gift code before.\n")
		c.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"message": "Sorry, you have already charged your wallet with this code before.",
		})
		return
	}

	url := "http://localhost:8002/getgift"
	fmt.Println("URL:>", url)
	code := GetGiftType{Code: req.GiftCode}
	jsonStr, err := json.Marshal(code)
	var resp PostReqReturnType
	resp = MakePostRequest(url, jsonStr, "application/json")
	// i, err := strconv.Atoi(resp.Status)
	
	if resp.Status != http.StatusOK {
		log.Printf("GiftCharge: Error on DiscountService returned status, Status: %v\n", resp.Status)
		c.JSON(resp.Status, gin.H{
			"status": resp.Status,
			"message":  "Something went wrong",
		})
		return
	}

	amount := resp.GiftAmount  // update
	insertError := dbConnect.Insert(&Transaction{
		Id: guuid.New().String(),
		Amount: amount,
		Type: false,
		Gift: true,
		GiftId: req.GiftCode,
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

	updateCash := user.Cash + resp.GiftAmount
	_, error1 := dbConnect.Model(&User{}).Set("cash = ?", updateCash).Where("id = ?", user.Id).Update()
	if error1 != nil {
		log.Printf("GiftCharge: Error, Reason: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": 500,
			"message":  "Something went wrong while updating user cash.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Wallet charged.",
	})
	return
}

// Who Gets Gift by code OK
func WhoGetsGift(c *gin.Context) {
	var req GiftChargeType
	c.BindJSON(&req)
	
	var user User
	err := dbConnect.Model(&user).Relation("Transactions",
		func(q *orm.Query) (*orm.Query, error) {
		return q.Where("gift_id = ?", req.GiftCode), nil}).Select()
	if err != nil {
		log.Printf("WhoGetsGift: Error while getting data from db, Reason: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusOK,
		"message": "Data gathered from Database successfully.",
		"data": user,
	})
	return
}

// Make Post Request based on Inputs OK
func MakePostRequest(url string, jsonStr []byte, contentType string) PostReqReturnType {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
        panic(err)
    }
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
    resp, err2 := client.Do(req)
    if err2 != nil {
        panic(err2)
    }
	defer resp.Body.Close()
	
	fmt.Println("response Status:", resp.Status)
    fmt.Println("response Headers:", resp.Header)
	
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	var res PostReqReturnType
	json.Unmarshal([]byte(body), &res)
	return res
}