package main

import (
	"fmt"
	"log"

	"github.com/avaneeshravat/school-management/exam-service/internal/config"
	"github.com/avaneeshravat/school-management/exam-service/internal/handler"
	"github.com/avaneeshravat/school-management/exam-service/internal/middleware"
	"github.com/avaneeshravat/school-management/exam-service/internal/model"
	"github.com/avaneeshravat/school-management/exam-service/internal/repository"
	"github.com/avaneeshravat/school-management/exam-service/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/avaneeshravat/school-management/exam-service/docs"
)

// @title           Exam Service API
// @version         1.0
// @description     Exam lifecycle service for exams, marks and results.
// @host            localhost:8086
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
	log.Println("connected to Exam DB")

	if err := db.AutoMigrate(&model.Exam{}, &model.Mark{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("exam database migrated")

	repo := repository.NewExamRepository(db)
	svc := service.NewExamService(repo)
	h := handler.NewExamHandler(svc)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "exam-service is running"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.POST("/exams", middleware.RequirePermission("create_exam"), h.CreateExam)
		protected.POST("/marks", middleware.RequirePermission("enter_marks"), h.EnterMarks)
		protected.POST("/results/publish", middleware.RequirePermission("publish_results"), h.PublishResults)
		protected.GET("/results/me", middleware.RequirePermission("view_own_results"), h.GetMyResults)
		protected.GET("/results", middleware.RequirePermission("view_results"), h.GetResults)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Exam Service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
