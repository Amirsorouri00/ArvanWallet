package main

import (
	"log"
	"github.com/gin-gonic/gin"

	db	"github.com/amirsorouri00/arvanwallet/db"
	routes "github.com/amirsorouri00/arvanwallet/routes"
)

func main() {
	// Connect to DB
	db.ConnectDB()

	// Initialize Router
	router := gin.Default()

	// Route Handlers / Endpoints
	routes.Routes(router)

	log.Fatal(router.Run(":8001"))
}