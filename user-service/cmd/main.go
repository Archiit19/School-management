package main

import (
	"fmt"
	"log"

	"github.com/Archiit19/School-management/user-service/internal/config"
	"github.com/Archiit19/School-management/user-service/internal/handler"
	"github.com/Archiit19/School-management/user-service/internal/middleware"
	"github.com/Archiit19/School-management/user-service/internal/model"
	"github.com/Archiit19/School-management/user-service/internal/repository"
	"github.com/Archiit19/School-management/user-service/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/Archiit19/School-management/user-service/docs"
)

// @title           User Service API
// @version         1.0
// @description     User profile CRUD for the School Management System.
// @host            localhost:8082
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in   header
// @name Authorization
func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("Connected to User DB")

	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("Database migrated")

	profileRepo, err := repository.NewProfileRepository(cfg)
	if err != nil {
		log.Fatalf("failed to connect to DynamoDB: %v", err)
	}
	log.Println("Connected to DynamoDB")

	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo, profileRepo, cfg)
	h := handler.NewUserHandler(svc)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "user-service is running"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	internal := r.Group("/internal")
	internal.Use(middleware.RequireInternalToken(cfg.InternalServiceToken))
	{
		internal.POST("/users", h.CreateProfileInternal)
		internal.GET("/users/by-email", h.GetUserByEmailInternal)
		internal.GET("/users/:id/profile", h.GetUserProfileInternal)
		internal.GET("/users/:id", h.GetUserInternal)
		internal.PATCH("/users/:id", h.UpdateProfileInternal)
		internal.DELETE("/users/:id", h.DeleteProfileInternal)
	}

	users := r.Group("/users")
	users.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		users.POST("", middleware.RequireAnyPermission("create_user", "admit_student"), h.CreateUser)
		users.GET("", middleware.RequireAnyPermission("view_users", "view_students", "admit_student"), h.GetUsers)
		users.GET("/me", middleware.RequirePermission("view_own_profile"), h.GetUserMe)
		users.GET("/:id", middleware.RequirePermission("view_users"), h.GetUserByID)
		users.PATCH("/:id", middleware.RequireAnyPermission("update_user", "update_student"), h.UpdateUser)
		users.DELETE("/:id", middleware.RequirePermission("delete_user"), h.DeleteUser)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("User Service starting on %s", addr)
	log.Printf("Swagger UI: http://localhost%s/swagger/index.html", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
