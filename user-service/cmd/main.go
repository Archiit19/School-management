package main

import (
	"fmt"
	"log"

	"github.com/avaneeshravat/school-management/user-service/internal/config"
	"github.com/avaneeshravat/school-management/user-service/internal/handler"
	"github.com/avaneeshravat/school-management/user-service/internal/middleware"
	"github.com/avaneeshravat/school-management/user-service/internal/model"
	"github.com/avaneeshravat/school-management/user-service/internal/repository"
	"github.com/avaneeshravat/school-management/user-service/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/avaneeshravat/school-management/user-service/docs" // swagger docs
)

// @title           User Service API
// @version         1.0
// @description     Roles & Permissions management service for the School Management System.
// @description     Handles role creation, permission management, and role-permission assignments.

// @host            localhost:8082
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
	log.Println("✅ Connected to User DB")

	// Auto-migrate
	if err := db.AutoMigrate(&model.Role{}, &model.Permission{}, &model.RolePermission{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("✅ Database migrated")

	// Wire dependencies
	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo)
	h := handler.NewUserHandler(svc)

	// Setup Gin router
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "user-service is running"})
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		// Internal endpoint (called by auth-service, no JWT needed)
		api.POST("/roles/internal", h.CreateRoleInternal)

		// Public read for role by ID (used by auth-service to fetch role name)
		api.GET("/roles/:id", h.GetRoleByID)
	}

	// Protected routes (require JWT)
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Roles
		protected.POST("/roles", h.CreateRole)
		protected.GET("/roles", h.GetRoles)
		protected.POST("/roles/assign-permission", h.AssignPermission)
		protected.GET("/roles/:id/permissions", h.GetRolePermissions)

		// Permissions
		protected.POST("/permissions", h.CreatePermission)
		protected.GET("/permissions", h.GetPermissions)
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 User Service starting on %s", addr)
	log.Printf("📖 Swagger UI: http://localhost%s/swagger/index.html", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
