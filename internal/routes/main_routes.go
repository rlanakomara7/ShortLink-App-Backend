package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/rlanakomara7/ShortLink-App-Backend/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(router *gin.Engine, db *pgxpool.Pool, jwtSecret string) {
	AuthRoutes(router, db, jwtSecret)
	LinkRoutes(router, db, jwtSecret)

	// Display Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
