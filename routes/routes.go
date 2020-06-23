package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	controller "github.com/amirsorouri00/arvanwallet/controller"
)

func Routes(router *gin.Engine) {
	router.GET("/", welcomeToWallet)
	
	// User APIs
	router.GET("/allusers", controller.GetAllUsers)
	router.POST("/seecash", controller.SeeWalletCash)
	router.POST("/adduser", controller.AddUser)

	// Transaction APIs
	router.GET("alltransactions", controller.GetAllTransactions)
	router.POST("/addtransaction", controller.AddTransaction)
	
	// Discount_Gift APIs
	router.POST("/giftcharge", controller.GiftCharge)
	router.POST("/whogetsgift", controller.WhoGetsGift)


	router.NoRoute(notFound)
}

func welcomeToWallet (c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": 200,
		"message": "Welcome to abrarvan wallet.",
	})
	return
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"status":  404,
		"message": "Route Not Found",
	})
	return
}