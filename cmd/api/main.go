package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Shortlink API is running",
		})
	})

	log.Println("Shortlink backend running on port 8080")

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
