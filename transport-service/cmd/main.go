package main

import (
	"fmt"
	"log"

	"github.com/avaneeshravat/school-management/transport-service/internal/config"
	"github.com/avaneeshravat/school-management/transport-service/internal/handler"
	"github.com/avaneeshravat/school-management/transport-service/internal/middleware"
	"github.com/avaneeshravat/school-management/transport-service/internal/model"
	"github.com/avaneeshravat/school-management/transport-service/internal/repository"
	"github.com/avaneeshravat/school-management/transport-service/internal/service"
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
	log.Println("connected to Transport DB")

	if err := db.AutoMigrate(
		&model.Vehicle{},
		&model.Route{},
		&model.Stop{},
		&model.StudentTransport{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("transport database migrated")

	// Create unique indexes to prevent duplicate records at DB level
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_vehicle_school_number ON vehicles (school_id, vehicle_number)`)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_route_school_code ON routes (school_id, route_code)`)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_student_transport_active ON student_transports (school_id, student_id) WHERE is_active = true`)
	log.Println("transport unique indexes ensured")

	repo := repository.NewTransportRepository(db)
	svc := service.NewTransportService(repo)
	h := handler.NewTransportHandler(svc)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "transport-service is running"})
	})

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Vehicles
		protected.POST("/vehicles", middleware.RequirePermission("manage_transport"), h.CreateVehicle)
		protected.GET("/vehicles", middleware.RequireAnyPermission("manage_transport", "view_transport"), h.GetVehicles)
		protected.GET("/vehicles/:id", middleware.RequireAnyPermission("manage_transport", "view_transport"), h.GetVehicle)
		protected.PATCH("/vehicles/:id", middleware.RequirePermission("manage_transport"), h.UpdateVehicle)
		protected.DELETE("/vehicles/:id", middleware.RequirePermission("manage_transport"), h.DeleteVehicle)

		// Routes
		protected.POST("/routes", middleware.RequirePermission("manage_transport"), h.CreateRoute)
		protected.GET("/routes", middleware.RequireAnyPermission("manage_transport", "view_transport"), h.GetRoutes)
		protected.GET("/routes/:id", middleware.RequireAnyPermission("manage_transport", "view_transport"), h.GetRoute)
		protected.PATCH("/routes/:id", middleware.RequirePermission("manage_transport"), h.UpdateRoute)
		protected.DELETE("/routes/:id", middleware.RequirePermission("manage_transport"), h.DeleteRoute)

		// Stops
		protected.POST("/stops", middleware.RequirePermission("manage_transport"), h.CreateStop)
		protected.GET("/stops", middleware.RequireAnyPermission("manage_transport", "view_transport"), h.GetStops)
		protected.GET("/stops/:id", middleware.RequireAnyPermission("manage_transport", "view_transport"), h.GetStop)
		protected.PATCH("/stops/:id", middleware.RequirePermission("manage_transport"), h.UpdateStop)
		protected.DELETE("/stops/:id", middleware.RequirePermission("manage_transport"), h.DeleteStop)

		// Student Transport Assignments
		protected.POST("/student-transport", middleware.RequirePermission("manage_transport"), h.CreateStudentTransport)
		protected.GET("/student-transport", middleware.RequireAnyPermission("manage_transport", "view_transport"), h.GetStudentTransports)
		protected.GET("/student-transport/:id", middleware.RequireAnyPermission("manage_transport", "view_transport"), h.GetStudentTransport)
		protected.PATCH("/student-transport/:id", middleware.RequirePermission("manage_transport"), h.UpdateStudentTransport)
		protected.DELETE("/student-transport/:id", middleware.RequirePermission("manage_transport"), h.DeleteStudentTransport)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Transport Service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
