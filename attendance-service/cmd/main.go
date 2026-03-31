package main

import (
	"fmt"
	"log"

	"github.com/avaneeshravat/school-management/attendance-service/internal/config"
	"github.com/avaneeshravat/school-management/attendance-service/internal/handler"
	"github.com/avaneeshravat/school-management/attendance-service/internal/middleware"
	"github.com/avaneeshravat/school-management/attendance-service/internal/model"
	"github.com/avaneeshravat/school-management/attendance-service/internal/repository"
	"github.com/avaneeshravat/school-management/attendance-service/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/avaneeshravat/school-management/attendance-service/docs"
)

// @title           Attendance Service API
// @version         1.0
// @description     Daily attendance management service.
// @host            localhost:8085
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
	log.Println("connected to Attendance DB")

	if err := db.AutoMigrate(&model.Attendance{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("attendance database migrated")

	repo := repository.NewAttendanceRepository(db)
	svc := service.NewAttendanceService(repo)
	h := handler.NewAttendanceHandler(svc)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "attendance-service is running"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.POST("/attendance", h.CreateAttendance)
		protected.GET("/attendance", h.GetAttendance)
		protected.PATCH("/attendance/:id", h.UpdateAttendance)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Attendance Service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
