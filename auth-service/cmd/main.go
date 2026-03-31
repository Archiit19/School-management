package main

import (
	"fmt"
	"log"

	"github.com/avaneeshravat/school-management/auth-service/internal/config"
	"github.com/avaneeshravat/school-management/auth-service/internal/handler"
	"github.com/avaneeshravat/school-management/auth-service/internal/middleware"
	"github.com/avaneeshravat/school-management/auth-service/internal/model"
	"github.com/avaneeshravat/school-management/auth-service/internal/repository"
	"github.com/avaneeshravat/school-management/auth-service/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/avaneeshravat/school-management/auth-service/docs" // swagger docs
)

// @title           Auth Service API
// @version         1.0
// @description     Authentication & User Management service for the School Management System.
// @description     Handles school registration, login, JWT authentication, and user CRUD operations.

// @host            localhost:8081
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in   header
// @name Authorization
// @description Enter your JWT token with the `Bearer ` prefix, e.g. `Bearer eyJhbGci...`

func main() {
	// Load config
	cfg := config.Load()

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("✅ Connected to Auth DB")

	// Auto-migrate
	if err := db.AutoMigrate(&model.School{}, &model.User{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("✅ Database migrated")

	// Wire dependencies
	repo := repository.NewAuthRepository(db)
	authSvc := service.NewAuthService(repo, cfg)
	userMgmtSvc := service.NewUserManagementService(repo, cfg, authSvc)

	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userMgmtSvc)

	// Setup Gin router
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "auth-service is running"})
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ─── Auth routes (public) ───────────────────────────────────────
	auth := r.Group("/auth")
	{
		auth.POST("/register-school", authHandler.RegisterSchool)
		auth.POST("/login", authHandler.Login)
	}

	// ─── Auth routes (protected) ────────────────────────────────────
	authProtected := r.Group("/auth")
	authProtected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		authProtected.GET("/me", authHandler.GetMe)
	}

	// ─── User Management routes (protected, admin only) ─────────────
	users := r.Group("/users")
	users.Use(middleware.JWTAuth(cfg.JWTSecret))
	users.Use(middleware.RequireRole("super_admin"))
	{
		users.POST("", userHandler.CreateUser)
		users.GET("", userHandler.GetUsers)
		users.GET("/:id", userHandler.GetUserByID)
		users.PATCH("/:id", userHandler.UpdateUser)
		users.DELETE("/:id", userHandler.DeleteUser)
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Auth Service starting on %s", addr)
	log.Printf("📖 Swagger UI: http://localhost%s/swagger/index.html", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
