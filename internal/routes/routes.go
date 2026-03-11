package routes

import (
	"database/sql"

	"github.com/Laiirton/time-control-api-go/internal/handlers"
	"github.com/Laiirton/time-control-api-go/internal/middleware"
	"github.com/Laiirton/time-control-api-go/internal/repository"
	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, db *sql.DB, jwtSecret string) {
	userRepo := repository.NewUserRepository(db)
	authHandler := handlers.NewAuthHandler(userRepo, jwtSecret)
	userHandler := handlers.NewUserHandler(userRepo)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			protected.GET("/auth/me", authHandler.Me)

			users := protected.Group("/users")
			{
				users.GET("", userHandler.List)
				users.GET("/:id", userHandler.GetByID)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
			}
		}
	}
}
