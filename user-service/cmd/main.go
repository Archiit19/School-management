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

var predefinedPermissions = []struct {
	Name        string
	Description string
}{
	{"create_user", "Create new users (staff, teacher, parent)"},
	{"view_users", "View the users list"},
	{"update_user", "Update user details"},
	{"delete_user", "Delete users"},
	{"create_role", "Create new roles"},
	{"manage_permissions", "Assign or revoke role permissions"},
	{"create_class", "Create classes"},
	{"create_section", "Create sections"},
	{"create_subject", "Create subjects"},
	{"view_academic", "View academic structure (classes, sections, subjects)"},
	{"admit_student", "Admit new students"},
	{"view_students", "View the students list"},
	{"update_student", "Update student details"},
	{"assign_teacher", "Assign teachers to class + subject"},
	{"mark_attendance", "Mark daily attendance"},
	{"view_attendance", "View attendance records"},
	{"mark_own_teacher_attendance", "Mark your own daily attendance as staff"},
	{"mark_teacher_attendance", "Mark any teacher or staff attendance"},
	{"view_teacher_attendance", "View teacher and staff attendance records"},
	{"create_assignment", "Create homework / assignments"},
	{"view_assignments", "View assignments"},
	{"submit_assignment", "Submit student work for assignments"},
	{"create_exam", "Create exams"},
	{"enter_marks", "Enter or update exam marks"},
	{"publish_results", "Publish exam results"},
	{"view_results", "View exam results"},
	{"create_fee", "Create fee structures"},
	{"record_payment", "Record fee payments"},
	{"view_dues", "View outstanding dues"},
}

func seedPermissions(db *gorm.DB) {
	for _, p := range predefinedPermissions {
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

	// Seed predefined permissions
	seedPermissions(db)
	log.Println("✅ Predefined permissions seeded")

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
		// Internal / public endpoints (called by auth-service, no JWT needed)
		api.POST("/roles/internal", h.CreateRoleInternal)
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
