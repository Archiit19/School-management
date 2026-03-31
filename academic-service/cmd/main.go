package main

import (
	"fmt"
	"log"

	"github.com/avaneeshravat/school-management/academic-service/internal/config"
	"github.com/avaneeshravat/school-management/academic-service/internal/handler"
	"github.com/avaneeshravat/school-management/academic-service/internal/middleware"
	"github.com/avaneeshravat/school-management/academic-service/internal/model"
	"github.com/avaneeshravat/school-management/academic-service/internal/repository"
	"github.com/avaneeshravat/school-management/academic-service/internal/service"
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
	log.Println("connected to Academic DB")

	if err := db.AutoMigrate(&model.Class{}, &model.Section{}, &model.Subject{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("academic database migrated")

	repo := repository.NewAcademicRepository(db)
	svc := service.NewAcademicService(repo)
	h := handler.NewAcademicHandler(svc)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "academic-service is running"})
	})

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.POST("/classes", h.CreateClass)
		protected.POST("/sections", h.CreateSection)
		protected.POST("/subjects", h.CreateSubject)
		protected.GET("/classes", h.GetClasses)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Academic Service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
