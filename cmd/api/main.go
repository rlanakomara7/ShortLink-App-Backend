package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/rlanakomara7/ShortLink-App-Backend/internal/config"
	"github.com/rlanakomara7/ShortLink-App-Backend/internal/database"
	"github.com/rlanakomara7/ShortLink-App-Backend/internal/routes"
)

// @title Shortlink App API
// @version 1.0
// @description REST API for Shortlink App
// @host localhost:8080
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	redisClient, err := database.ConnectRedis()
	if err != nil {
		log.Fatal("Failed to connect Redis :", err)
	}
	defer redisClient.Close()
	log.Println("Connected to Redis")

	router := gin.Default()

	routes.Setup(router, db, cfg.JWTSecret, redisClient)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
