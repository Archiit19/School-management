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
		protected.POST("/classes", h.CreateClass)
		protected.POST("/sections", h.CreateSection)
		protected.POST("/subjects", h.CreateSubject)
		protected.GET("/classes", h.GetClasses)
		protected.POST("/teacher-assignments", h.CreateTeacherAssignment)
		protected.GET("/teacher-assignments", h.GetTeacherAssignments)
		protected.POST("/assignments", h.CreateAssignment)
		protected.GET("/assignments", h.GetAssignments)
		protected.POST("/submissions", h.CreateSubmission)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Academic Service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
