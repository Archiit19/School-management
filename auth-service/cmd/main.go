package main

import (
	"fmt"
	"log"

	"github.com/Archiit19/School-management/auth-service/internal/config"
	"github.com/Archiit19/School-management/auth-service/internal/handler"
	"github.com/Archiit19/School-management/pkg/middleware"
	"github.com/Archiit19/School-management/auth-service/internal/migrate"
	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/Archiit19/School-management/auth-service/internal/rbacdata"
	"github.com/Archiit19/School-management/auth-service/internal/repository"
	"github.com/Archiit19/School-management/auth-service/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/Archiit19/School-management/auth-service/docs"
)

// @title           Auth Service API
// @version         1.0
// @description     Authentication, credentials, RBAC, and JWT for the School Management System.
// @host            localhost:8081
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in   header
// @name Authorization
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
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("Connected to Auth DB")

	if err := db.AutoMigrate(
		&model.UserCredential{},
		&model.UserRole{},
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.RoleField{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	if err := migrate.LegacySchema(db); err != nil {
		log.Fatalf("failed legacy migration: %v", err)
	}
	log.Println("Database migrated")

	seedPermissionsFromJSON(db)
	log.Println("Predefined permissions seeded")

	rbacRepo := repository.NewRBACRepository(db)
	credRepo := repository.NewCredentialRepository(db)
	rbacSvc := service.NewRBACService(rbacRepo)
	credSvc := service.NewCredentialService(credRepo, rbacSvc)
	authSvc := service.NewAuthService(cfg, credSvc, rbacSvc)

	if err := rbacSvc.SyncTemplateRolesForAllSchools(); err != nil {
		log.Printf("template role sync: %v", err)
	}

	authHandler := handler.NewAuthHandler(authSvc)
	rbacHandler := handler.NewRBACHandler(rbacSvc)
	internalHandler := handler.NewInternalHandler(credSvc, rbacSvc)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "auth-service is running"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	auth := r.Group("/auth")
	{
		auth.POST("/signup", authHandler.Signup)
		auth.POST("/register-school", authHandler.RegisterSchool)
		auth.POST("/login", authHandler.Login)
	}

	jwtAuth := middleware.JWTAuth(cfg.JWTSecret,
		middleware.WithUUIDClaim("role_id"),
		middleware.WithStringClaim("email"),
	)

	authProtected := r.Group("/auth")
	authProtected.Use(jwtAuth)
	{
		authProtected.GET("/me", authHandler.GetMe)
		authProtected.PATCH("/me", authHandler.UpdateProfile)
		authProtected.POST("/select-school", authHandler.SelectSchool)
		authProtected.POST("/exit-school", authHandler.ExitSchool)
	}

	api := r.Group("/api/v1")
	{
		api.POST("/roles/internal", rbacHandler.CreateRoleInternal)
		api.POST("/internal/bootstrap-school", rbacHandler.BootstrapSchoolInternal)
		api.GET("/internal/roles/by-name", rbacHandler.GetRoleByNameAndSchoolInternal)
		api.GET("/roles/:id", rbacHandler.GetRoleByID)
		api.GET("/roles/:id/permissions", rbacHandler.GetRolePermissions)
		api.GET("/roles/:id/fields", rbacHandler.GetRoleFields)
	}

	protected := api.Group("")
	protected.Use(jwtAuth)
	{
		protected.POST("/roles", middleware.RequirePermission("create_role"), rbacHandler.CreateRole)
		protected.GET("/roles", rbacHandler.GetRoles)
		protected.PUT("/roles/:id/fields", middleware.RequirePermission("create_role"), rbacHandler.UpdateRoleFields)
		protected.POST("/roles/assign-permission", middleware.RequirePermission("manage_permissions"), rbacHandler.AssignPermission)
		protected.DELETE("/roles/:id/permissions/:permissionId", middleware.RequirePermission("manage_permissions"), rbacHandler.RemovePermissionFromRole)
		protected.GET("/permissions", rbacHandler.GetPermissions)
	}

	internal := r.Group("/internal")
	internal.Use(middleware.RequireInternalToken(cfg.InternalServiceToken))
	{
		internal.POST("/credentials", internalHandler.SetCredential)
		internal.DELETE("/credentials/:userId", internalHandler.DeleteCredential)
		internal.POST("/user-roles", internalHandler.AssignUserRole)
		internal.PATCH("/user-roles", internalHandler.UpdateUserRole)
		internal.DELETE("/user-roles", internalHandler.RemoveUserRole)
		internal.GET("/user-roles/:userId", internalHandler.ListUserRoles)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Auth Service starting on %s", addr)
	log.Printf("Swagger UI: http://localhost%s/swagger/index.html", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
