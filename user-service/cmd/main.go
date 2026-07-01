package main

import (
	"context"
	"time"

	log "github.com/Archiit19/School-management/pkg/logger"
	pkgconfig "github.com/Archiit19/School-management/pkg/config"
	"github.com/Archiit19/School-management/pkg/health"
	"github.com/Archiit19/School-management/pkg/middleware"
	"github.com/Archiit19/School-management/pkg/server"
	"github.com/Archiit19/School-management/pkg/tracer"
	"github.com/Archiit19/School-management/user-service/internal/config"
	"github.com/Archiit19/School-management/user-service/internal/handler"
	"github.com/Archiit19/School-management/user-service/internal/model"
	"github.com/Archiit19/School-management/user-service/internal/repository"
	"github.com/Archiit19/School-management/user-service/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	if _, err := log.InitFromEnv("user-service"); err != nil {
		log.Fatal("failed to initialize logger", log.Err(err))
	}

	traceShutdown, err := tracer.InitFromEnv("user-service")
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

	db, err := pkgconfig.OpenGORM(cfg.DSN(), &pkgconfig.GORMOptions{
		Config: &gorm.Config{Logger: gormlogger.Discard},
	})
	if err != nil {
		log.Fatal("failed to connect to database", log.Err(err))
	}
	if err := tracer.InstrumentGORM(db); err != nil {
		log.Fatal("failed to instrument database", log.Err(err))
	}
	log.Info("connected to database")

	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatal("failed to migrate database", log.Err(err))
	}
	log.Info("database migrated")

	profileRepo, err := repository.NewProfileRepository(cfg)
	if err != nil {
		log.Fatal("failed to connect to dynamodb", log.Err(err))
	}
	log.Info("connected to dynamodb")

	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo, profileRepo, cfg)
	h := handler.NewUserHandler(svc)

	r := middleware.NewEngine("user-service")
	health.Register(r, "user-service", health.CheckDB(db))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	internal := r.Group("/internal")
	internal.Use(middleware.RequireInternalToken(cfg.InternalServiceToken))
	{
		internal.POST("/users", h.CreateProfileInternal)
		internal.GET("/users/by-email", h.GetUserByEmailInternal)
		internal.GET("/users/has-child/:parentId/:childId", h.ParentHasChildInternal)
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
		users.GET("/me/children", middleware.RequirePermission("view_own_profile"), h.GetMyChildren)
		users.GET("/me/children/:childId", middleware.RequirePermission("view_own_profile"), h.GetChildForParent)
		users.GET("/:id", middleware.RequirePermission("view_users"), h.GetUserByID)
		users.PATCH("/:id", middleware.RequireAnyPermission("update_user", "update_student"), h.UpdateUser)
		users.DELETE("/:id", middleware.RequirePermission("delete_user"), h.DeleteUser)
	}

	if err := server.Run(r, server.LoadConfigFromEnv(cfg.Port)); err != nil {
		log.Fatal("failed to start server", log.Err(err))
	}
}
