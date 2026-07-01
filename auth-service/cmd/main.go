package main

import (
	"context"
	"time"

	log "github.com/Archiit19/School-management/pkg/logger"
	pkgconfig "github.com/Archiit19/School-management/pkg/config"
	"github.com/Archiit19/School-management/auth-service/internal/config"
	"github.com/Archiit19/School-management/auth-service/internal/handler"
	"github.com/Archiit19/School-management/pkg/health"
	"github.com/Archiit19/School-management/pkg/middleware"
	"github.com/Archiit19/School-management/auth-service/internal/migrate"
	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/Archiit19/School-management/auth-service/internal/rbacdata"
	"github.com/Archiit19/School-management/auth-service/internal/repository"
	"github.com/Archiit19/School-management/auth-service/internal/service"
	"github.com/Archiit19/School-management/pkg/server"
	"github.com/Archiit19/School-management/pkg/tracer"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

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
		log.Fatal("failed to load predefined_permissions.json", log.Err(err))
	}
	for _, p := range list {
		db.Where("name = ?", p.Name).FirstOrCreate(&model.Permission{
			Name:        p.Name,
			Description: p.Description,
		})
	}
}

func main() {
	if _, err := log.InitFromEnv("auth-service"); err != nil {
		log.Fatal("failed to initialize logger", log.Err(err))
	}

	traceShutdown, err := tracer.InitFromEnv("auth-service")
	if err != nil {
		log.Fatal("failed to initialize tracer", log.Err(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := traceShutdown(ctx); err != nil {
			log.Error("tracer shutdown", log.Err(err))
		}
	}()

	cfg := config.Load()
	if err := pkgconfig.ValidateCommon(cfg.JWTSecret, cfg.InternalServiceToken); err != nil {
		log.Fatal("invalid configuration", log.Err(err))
	}

	db, err := pkgconfig.OpenGORM(cfg.DSN(), nil)
	if err != nil {
		log.Fatal("failed to connect to database", log.Err(err))
	}
	if err := tracer.InstrumentGORM(db); err != nil {
		log.Fatal("failed to instrument database", log.Err(err))
	}
	log.Info("connected to database")

	if err := db.AutoMigrate(
		&model.UserCredential{},
		&model.UserRole{},
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.RoleField{},
	); err != nil {
		log.Fatal("failed to migrate database", log.Err(err))
	}
	if err := migrate.LegacySchema(db); err != nil {
		log.Fatal("failed legacy migration", log.Err(err))
	}
	log.Info("database migrated")

	seedPermissionsFromJSON(db)
	log.Info("predefined permissions seeded")

	rbacRepo := repository.NewRBACRepository(db)
	credRepo := repository.NewCredentialRepository(db)
	rbacSvc := service.NewRBACService(rbacRepo)
	credSvc := service.NewCredentialService(credRepo, rbacSvc)
	authSvc := service.NewAuthService(cfg, credSvc, rbacSvc)

	if err := rbacSvc.SyncTemplateRolesForAllSchools(); err != nil {
		log.Warn("template role sync failed", log.Err(err))
	}

	authHandler := handler.NewAuthHandler(authSvc)
	rbacHandler := handler.NewRBACHandler(rbacSvc)
	internalHandler := handler.NewInternalHandler(credSvc, rbacSvc)

	r := middleware.NewEngine("auth-service")
	health.Register(r, "auth-service", health.CheckDB(db))

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

	if err := server.Run(r, server.LoadConfigFromEnv(cfg.Port)); err != nil {
		log.Fatal("failed to start server", log.Err(err))
	}
}
