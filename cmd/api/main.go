package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/rlanakomara7/ShortLink-App-Backend/internal/config"
	"github.com/rlanakomara7/ShortLink-App-Backend/internal/database"
	"github.com/rlanakomara7/ShortLink-App-Backend/internal/routes"
)

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

	router := gin.Default()

	routes.Setup(router, db, cfg.JWTSecret)

	log.Println("Database connected successfully")

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
