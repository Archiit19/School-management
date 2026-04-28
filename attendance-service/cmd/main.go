package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

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

	if err := db.AutoMigrate(&model.Attendance{}, &model.TeacherAttendance{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("attendance database migrated")

	if err := db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS ux_teacher_attendance_school_teacher_date
ON teacher_attendances (school_id, teacher_user_id, date);
`).Error; err != nil {
		log.Printf("warn: teacher attendance unique index: %v", err)
	}
	if err := db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS ux_student_attendance_scope
ON attendances (
	school_id,
	student_id,
	class_id,
	date,
	COALESCE(section_id, '00000000-0000-0000-0000-000000000000'::uuid),
	COALESCE(subject_id, '00000000-0000-0000-0000-000000000000'::uuid)
);
`).Error; err != nil {
		log.Printf("warn: student attendance unique index: %v", err)
	}

	repo := repository.NewAttendanceRepository(db)
	httpClient := &http.Client{Timeout: 8 * time.Second}
	svc := service.NewAttendanceService(repo, cfg, httpClient)
	h := handler.NewAttendanceHandler(svc)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "attendance-service is running"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.POST("/attendance", middleware.RequirePermission("mark_attendance"), h.CreateAttendance)
		protected.POST("/attendance/bulk", middleware.RequirePermission("mark_attendance"), h.BulkCreateAttendance)
		protected.GET("/attendance/me", middleware.RequirePermission("view_own_attendance"), h.GetMyAttendance)
		protected.GET("/attendance/me/stats", middleware.RequirePermission("view_own_attendance"), h.GetMyAttendanceStats)
		protected.GET("/attendance", middleware.RequirePermission("view_attendance"), h.GetAttendance)
		protected.GET("/attendance/stats", middleware.RequirePermission("view_attendance"), h.GetAttendanceStats)
		protected.PATCH("/attendance/:id", middleware.RequirePermission("mark_attendance"), h.UpdateAttendance)

		protected.POST("/teacher-attendance", middleware.RequireAnyPermission("mark_teacher_attendance", "mark_own_teacher_attendance"), h.CreateTeacherAttendance)
		protected.POST("/teacher-attendance/bulk", middleware.RequirePermission("mark_teacher_attendance"), h.BulkCreateTeacherAttendance)
		protected.GET("/teacher-attendance", middleware.RequireAnyPermission("view_teacher_attendance", "mark_teacher_attendance", "mark_own_teacher_attendance"), h.GetTeacherAttendance)
		protected.GET("/teacher-attendance/stats", middleware.RequireAnyPermission("view_teacher_attendance", "mark_teacher_attendance", "mark_own_teacher_attendance"), h.GetTeacherAttendanceStats)
		protected.PATCH("/teacher-attendance/:id", middleware.RequireAnyPermission("mark_teacher_attendance", "mark_own_teacher_attendance"), h.UpdateTeacherAttendance)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Attendance Service starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
