package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	_ "github.com/rlanakomara7/ShortLink-App-Backend/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(router *gin.Engine, db *pgxpool.Pool, jwtSecret string, redisClient *redis.Client) {
	AuthRoutes(router, db, jwtSecret, redisClient)
	LinkRoutes(router, db, jwtSecret, redisClient)

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
