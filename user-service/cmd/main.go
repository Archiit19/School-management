package main

import (
	"fmt"

	log "github.com/Archiit19/School-management/pkg/logger"
	"github.com/Archiit19/School-management/pkg/middleware"
	"github.com/Archiit19/School-management/user-service/internal/config"
	"github.com/Archiit19/School-management/user-service/internal/handler"
	"github.com/Archiit19/School-management/user-service/internal/model"
	"github.com/Archiit19/School-management/user-service/internal/repository"
	"github.com/Archiit19/School-management/user-service/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

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

	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Discard,
	})
	if err != nil {
		log.Fatal("failed to connect to database", log.Err(err))
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

	r := middleware.NewEngine()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "user-service is running"})
	})
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

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Info("starting http server",
		log.AddField("addr", addr),
		log.AddField("swagger", fmt.Sprintf("http://localhost%s/swagger/index.html", addr)),
	)
	if err := r.Run(addr); err != nil {
		log.Fatal("failed to start server", log.Err(err))
	}
}
