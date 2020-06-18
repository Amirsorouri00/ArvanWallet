package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	//controller "./controller"
)

func Routes(router *gin.Engine) {
	router.GET("/", welcomeToWallet)
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