package main

import (
	log "github.com/Archiit19/School-management/pkg/logger"
	pkgconfig "github.com/Archiit19/School-management/pkg/config"
	"github.com/Archiit19/School-management/attendance-service/internal/config"
	"github.com/Archiit19/School-management/attendance-service/internal/handler"
	"github.com/Archiit19/School-management/pkg/health"
	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/Archiit19/School-management/pkg/middleware"
	"github.com/Archiit19/School-management/pkg/server"
	"github.com/Archiit19/School-management/pkg/userclient"
	"github.com/Archiit19/School-management/attendance-service/internal/model"
	"github.com/Archiit19/School-management/attendance-service/internal/repository"
	"github.com/Archiit19/School-management/attendance-service/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Archiit19/School-management/attendance-service/docs"
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
	if _, err := log.InitFromEnv("attendance-service"); err != nil {
		log.Fatal("failed to initialize logger", log.Err(err))
	}

	cfg := config.Load()
	if err := pkgconfig.ValidateCommon(cfg.JWTSecret, cfg.InternalServiceToken); err != nil {
		log.Fatal("invalid configuration", log.Err(err))
	}

	db, err := pkgconfig.OpenGORM(cfg.DSN(), nil)
	if err != nil {
		log.Fatal("failed to connect to database", log.Err(err))
	}
	log.Info("connected to database")

	if err := db.AutoMigrate(&model.Attendance{}, &model.TeacherAttendance{}); err != nil {
		log.Fatal("failed to migrate database", log.Err(err))
	}
	log.Info("database migrated")

	if err := db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS ux_teacher_attendance_school_teacher_date
ON teacher_attendances (school_id, teacher_user_id, date);
`).Error; err != nil {
		log.Warn("teacher attendance unique index", log.Err(err))
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
		log.Warn("student attendance unique index", log.Err(err))
	}

	repo := repository.NewAttendanceRepository(db)
	httpCfg := pkgconfig.LoadHTTPClientConfigFromEnv()
	userInternal := httpclient.NewFromConfig(httpclient.ClientConfig{
		BaseURL: cfg.UserServiceURL,
		Token:   cfg.InternalServiceToken,
		Name:    "user-service",
		HTTP:    &httpCfg,
	})
	outbound := httpclient.OutboundHTTP("outbound")
	svc := service.NewAttendanceService(repo, cfg, userInternal, outbound)
	users := userclient.New(cfg.UserServiceURL, cfg.InternalServiceToken)
	h := handler.NewAttendanceHandler(svc, users)

	r := middleware.NewEngine()
	health.Register(r, "attendance-service", health.CheckDB(db))

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

	if err := server.Run(r, server.LoadConfigFromEnv(cfg.Port)); err != nil {
		log.Fatal("failed to start server", log.Err(err))
	}
}
