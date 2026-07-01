package main

import (
	"context"
	"time"

	log "github.com/Archiit19/School-management/pkg/logger"
	pkgconfig "github.com/Archiit19/School-management/pkg/config"
	"github.com/Archiit19/School-management/academic-service/internal/config"
	"github.com/Archiit19/School-management/academic-service/internal/handler"
	"github.com/Archiit19/School-management/pkg/health"
	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/Archiit19/School-management/pkg/middleware"
	"github.com/Archiit19/School-management/pkg/server"
	"github.com/Archiit19/School-management/pkg/tracer"
	"github.com/Archiit19/School-management/pkg/userclient"
	"github.com/Archiit19/School-management/academic-service/internal/model"
	"github.com/Archiit19/School-management/academic-service/internal/repository"
	"github.com/Archiit19/School-management/academic-service/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Archiit19/School-management/academic-service/docs"
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
	if _, err := log.InitFromEnv("academic-service"); err != nil {
		log.Fatal("failed to initialize logger", log.Err(err))
	}

	traceShutdown, err := tracer.InitFromEnv("academic-service")
	if err != nil {
		log.Fatal("failed to initialize tracer", log.Err(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := traceShutdown(ctx); err != nil {
			log.Error("tracer shutdown", log.Err(err))
		}
	}()

	cfg := config.Load()
	if err := pkgconfig.ValidateCommon(cfg.JWTSecret, cfg.InternalServiceToken); err != nil {
		log.Fatal("invalid configuration", log.Err(err))
	}

	db, err := pkgconfig.OpenGORM(cfg.DSN(), nil)
	if err != nil {
		log.Fatal("failed to connect to database", log.Err(err))
	}
	if err := tracer.InstrumentGORM(db); err != nil {
		log.Fatal("failed to instrument database", log.Err(err))
	}
	log.Info("connected to database")

	if err := db.AutoMigrate(
		&model.Class{},
		&model.Section{},
		&model.Subject{},
		&model.TeacherAssignment{},
		&model.Assignment{},
		&model.Submission{},
		&model.StudentEnrollment{},
	); err != nil {
		log.Fatal("failed to migrate database", log.Err(err))
	}
	log.Info("database migrated")

	repo := repository.NewAcademicRepository(db)
	httpCfg := pkgconfig.LoadHTTPClientConfigFromEnv()
	userInternal := httpclient.NewFromConfig(httpclient.ClientConfig{
		BaseURL: cfg.UserServiceURL,
		Token:   cfg.InternalServiceToken,
		Name:    "user-service",
		HTTP:    &httpCfg,
	})
	outbound := httpclient.OutboundHTTP("outbound")
	svc := service.NewAcademicService(repo, cfg, userInternal, outbound)
	users := userclient.New(cfg.UserServiceURL, cfg.InternalServiceToken)
	h := handler.NewAcademicHandler(svc, users)
	eh := handler.NewEnrollmentHandler(svc, users)

	r := middleware.NewEngine("academic-service")
	health.Register(r, "academic-service", health.CheckDB(db))

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
		protected.PATCH("/teacher-assignments/:id", middleware.RequirePermission("assign_teacher"), h.UpdateTeacherAssignment)
		protected.DELETE("/teacher-assignments/:id", middleware.RequirePermission("assign_teacher"), h.DeleteTeacherAssignment)
		protected.GET("/academic/me", middleware.RequirePermission("view_own_profile"), h.GetMyAcademicProfile)
		protected.GET("/enrollments/me", middleware.RequirePermission("view_own_profile"), eh.GetMyEnrollment)
		protected.GET("/enrollments", middleware.RequireAnyPermission("view_students", "mark_attendance", "view_academic", "enter_marks"), eh.ListEnrollments)
		protected.POST("/assignments", middleware.RequirePermission("create_assignment"), h.CreateAssignment)
		protected.GET("/assignments/me", middleware.RequirePermission("view_own_assignments"), h.GetMyAssignments)
		protected.GET("/assignments", middleware.RequirePermission("view_assignments"), h.GetAssignments)
		protected.GET("/assignments/:id/submissions", middleware.RequirePermission("view_assignments"), h.GetAssignmentSubmissions)
		protected.GET("/submissions/me", middleware.RequirePermission("view_own_submissions"), h.GetMySubmissions)
		protected.POST("/submissions/me", middleware.RequirePermission("submit_own_assignment"), h.CreateMySubmission)
		protected.POST("/submissions", middleware.RequirePermission("submit_assignment"), h.CreateSubmission)
		protected.PATCH("/submissions/:id", middleware.RequirePermission("create_assignment"), h.ReviewSubmission)
	}

	internal := r.Group("/internal")
	internal.Use(middleware.RequireInternalToken(cfg.InternalServiceToken))
	{
		internal.POST("/enrollments", eh.UpsertEnrollmentInternal)
		internal.GET("/enrollments/:userId", eh.GetEnrollmentInternal)
		internal.DELETE("/enrollments/:userId", eh.DeleteEnrollmentInternal)
	}

	if err := server.Run(r, server.LoadConfigFromEnv(cfg.Port)); err != nil {
		log.Fatal("failed to start server", log.Err(err))
	}
}
