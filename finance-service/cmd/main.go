package main

import (
	"fmt"
	"log"

	"github.com/avaneeshravat/school-management/finance-service/internal/config"
	"github.com/avaneeshravat/school-management/finance-service/internal/handler"
	"github.com/avaneeshravat/school-management/finance-service/internal/middleware"
	"github.com/avaneeshravat/school-management/finance-service/internal/model"
	"github.com/avaneeshravat/school-management/finance-service/internal/repository"
	"github.com/avaneeshravat/school-management/finance-service/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/avaneeshravat/school-management/finance-service/docs"
)

// @title           Finance Service API
// @version         1.0
// @description     Fee structures, payments, and dues management.
// @host            localhost:8087
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
	log.Println("connected to Finance DB")

	if err := db.AutoMigrate(&model.Fee{}, &model.Payment{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("finance database migrated")

	repo := repository.NewFinanceRepository(db)
	svc := service.NewFinanceService(repo)
	h := handler.NewFinanceHandler(svc)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "finance-service is running"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.POST("/fees", h.CreateFee)
		protected.POST("/payments", h.RecordPayment)
		protected.GET("/dues", h.GetDues)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Finance Service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
