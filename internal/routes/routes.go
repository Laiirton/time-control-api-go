package routes

import (
	"database/sql"

	"github.com/Laiirton/time-control-api-go/internal/config"
	"github.com/Laiirton/time-control-api-go/internal/handlers"
	"github.com/Laiirton/time-control-api-go/internal/middleware"
	"github.com/Laiirton/time-control-api-go/internal/repository"
	"github.com/Laiirton/time-control-api-go/internal/storage"
	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, db *sql.DB, cfg *config.Config) {
	userRepo := repository.NewUserRepository(db)
	timeRecordRepo := repository.NewTimeRecordRepository(db)
	storageClient := storage.NewSupabaseStorage(cfg.SupabaseURL, cfg.SupabaseServiceRoleKey, cfg.SupabaseStorageBucket)

	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret)
	userHandler := handlers.NewUserHandler(userRepo)
	timeRecordHandler := handlers.NewTimeRecordHandler(timeRecordRepo, storageClient)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.GET("/auth/me", authHandler.Me)
			protected.POST("/time-records/clock", timeRecordHandler.Clock)
			protected.GET("/time-records/me", timeRecordHandler.Me)
			protected.GET("/time-records/me/today", timeRecordHandler.MeToday)
			protected.GET("/time-records", timeRecordHandler.List)

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
