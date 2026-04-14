package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/avaneeshravat/school-management/academic-service/internal/config"
	"github.com/avaneeshravat/school-management/academic-service/internal/handler"
	"github.com/avaneeshravat/school-management/academic-service/internal/middleware"
	"github.com/avaneeshravat/school-management/academic-service/internal/model"
	"github.com/avaneeshravat/school-management/academic-service/internal/repository"
	"github.com/avaneeshravat/school-management/academic-service/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/avaneeshravat/school-management/academic-service/docs"
)

// @title           Academic Service API
// @version         1.0
// @description     Academic management service for school structure, assignments, and submissions.
// @host            localhost:8083
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
	log.Println("connected to Academic DB")

	if err := db.AutoMigrate(
		&model.Class{},
		&model.Section{},
		&model.Subject{},
		&model.TeacherAssignment{},
		&model.Assignment{},
		&model.Submission{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("academic database migrated")

	repo := repository.NewAcademicRepository(db)
	httpClient := &http.Client{Timeout: 8 * time.Second}
	svc := service.NewAcademicService(repo, cfg, httpClient)
	h := handler.NewAcademicHandler(svc)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "academic-service is running"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.POST("/classes", middleware.RequirePermission("create_class"), h.CreateClass)
		protected.POST("/sections", middleware.RequirePermission("create_section"), h.CreateSection)
		protected.POST("/subjects", middleware.RequirePermission("create_subject"), h.CreateSubject)
		protected.GET("/classes", middleware.RequirePermission("view_academic"), h.GetClasses)
		protected.POST("/teacher-assignments", middleware.RequirePermission("assign_teacher"), h.CreateTeacherAssignment)
		protected.GET("/teacher-assignments", middleware.RequirePermission("view_academic"), h.GetTeacherAssignments)
		protected.POST("/assignments", middleware.RequirePermission("create_assignment"), h.CreateAssignment)
		protected.GET("/assignments", middleware.RequirePermission("view_assignments"), h.GetAssignments)
		protected.POST("/submissions", middleware.RequirePermission("submit_assignment"), h.CreateSubmission)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Academic Service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
