package main

import (
	"fmt"
	"log"

	"github.com/avaneeshravat/school-management/user-service/internal/config"
	"github.com/avaneeshravat/school-management/user-service/internal/handler"
	"github.com/avaneeshravat/school-management/user-service/internal/middleware"
	"github.com/avaneeshravat/school-management/user-service/internal/model"
	"github.com/avaneeshravat/school-management/user-service/internal/rbacdata"
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

func seedPermissionsFromJSON(db *gorm.DB) {
	list, err := rbacdata.LoadPredefinedPermissions()
	if err != nil {
		log.Fatalf("failed to load predefined_permissions.json: %v", err)
	}
	for _, p := range list {
		db.Where("name = ?", p.Name).FirstOrCreate(&model.Permission{
			Name:        p.Name,
			Description: p.Description,
		})
	}
}

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

	// Seed predefined permissions from internal/rbacdata/predefined_permissions.json
	seedPermissionsFromJSON(db)
	log.Println("✅ Predefined permissions seeded from JSON")

	// Wire dependencies
	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo)
	h := handler.NewUserHandler(svc)

	// Ensure template roles exist for schools already in DB (creates missing roles + syncs permission links).
	if err := svc.SyncTemplateRolesForAllSchools(); err != nil {
		log.Printf("⚠️ template role sync (existing schools): %v", err)
	} else {
		log.Println("✅ Template roles synced for existing schools")
	}

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
		// Internal / public endpoints (called by auth-service, no JWT needed)
		api.POST("/roles/internal", h.CreateRoleInternal)
		api.POST("/internal/bootstrap-school", h.BootstrapSchoolInternal)
		api.GET("/roles/:id", h.GetRoleByID)
		api.GET("/roles/:id/permissions", h.GetRolePermissions)
	}

	// Protected routes (require JWT)
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.POST("/roles", middleware.RequirePermission("create_role"), h.CreateRole)
		protected.GET("/roles", h.GetRoles)
		protected.POST("/roles/assign-permission", middleware.RequirePermission("manage_permissions"), h.AssignPermission)

		// Permissions (predefined, read-only)
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
