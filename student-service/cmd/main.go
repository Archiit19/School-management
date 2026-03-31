package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/avaneeshravat/school-management/student-service/internal/config"
	"github.com/avaneeshravat/school-management/student-service/internal/handler"
	"github.com/avaneeshravat/school-management/student-service/internal/middleware"
	"github.com/avaneeshravat/school-management/student-service/internal/model"
	"github.com/avaneeshravat/school-management/student-service/internal/repository"
	"github.com/avaneeshravat/school-management/student-service/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("connected to Student DB")

	if err := db.AutoMigrate(&model.Student{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("student database migrated")

	repo := repository.NewStudentRepository(db)
	httpClient := &http.Client{Timeout: 8 * time.Second}
	svc := service.NewStudentService(repo, cfg, httpClient)
	h := handler.NewStudentHandler(svc)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "student-service is running"})
	})

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.POST("/students", h.CreateStudent)
		protected.GET("/students", h.GetStudents)
		protected.PATCH("/students/:id", h.UpdateStudent)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Student Service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
