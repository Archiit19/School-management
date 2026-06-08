package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Archiit19/School-management/school-service/internal/config"
	"github.com/Archiit19/School-management/school-service/internal/handler"
	"github.com/Archiit19/School-management/school-service/internal/middleware"
	"github.com/Archiit19/School-management/school-service/internal/model"
	"github.com/Archiit19/School-management/school-service/internal/repository"
	"github.com/Archiit19/School-management/school-service/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title           School Service API
// @version         1.0
// @description     School registry with admin mapping and CRUD.
// @host            localhost:8088
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("connected to School DB")

	if err := db.AutoMigrate(&model.School{}, &model.UserSchool{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("school database migrated")

	repo := repository.NewSchoolRepository(db)
	svc := service.NewSchoolService(repo, cfg.UserServiceURL)
	h := handler.NewSchoolHandler(svc)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "school-service is running"})
	})

	internal := r.Group("/internal")
	internal.Use(middleware.RequireInternalToken(cfg.InternalServiceToken))
	{
		internal.POST("/schools/with-admin", h.CreateSchoolWithAdminInternal)
		internal.GET("/schools/by-email", h.GetSchoolByEmailInternal)
		internal.GET("/schools/by-user/:userId", h.ListSchoolsByUserInternal)
		internal.GET("/users/:userId/memberships", h.ListMembershipsForUserInternal)
		internal.GET("/schools/:id/members", h.ListMembersForSchoolInternal)
		internal.POST("/schools/:id/members", h.AddMemberInternal)
		internal.GET("/schools/:id/members/:userId", h.GetMemberInternal)
		internal.PATCH("/schools/:id/members/:userId", h.UpdateMemberInternal)
		internal.DELETE("/schools/:id/members/:userId", h.RemoveMemberInternal)
		internal.GET("/schools/:id/admins/:userId", h.CheckAdminInternal)
		internal.GET("/schools/:id", h.GetSchoolInternal)
	}

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.GET("/schools/mine", middleware.RequirePermission("view_my_schools"), h.ListMySchools)
		protected.POST("/schools", middleware.RequirePermission("create_school"), h.CreateSchool)
		protected.GET("/schools", middleware.RequirePermission("view_all_schools"), h.ListSchools)
		protected.GET("/schools/me", middleware.RequirePermission("view_school"), h.GetMySchool)
		protected.PATCH("/schools/me", middleware.RequirePermission("manage_school"), h.UpdateMySchool)

		adminSchool := protected.Group("/schools/:id")
		adminSchool.Use(middleware.RequireSchoolAdmin(svc))
		{
			adminSchool.GET("", middleware.RequirePermission("view_school"), h.GetSchool)
			adminSchool.PATCH("", middleware.RequirePermission("manage_school"), h.UpdateSchool)
			adminSchool.DELETE("", middleware.RequirePermission("manage_school"), h.DeleteSchool)
		}
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("School Service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
