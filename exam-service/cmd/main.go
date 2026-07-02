package main

import (
	"context"
	"time"

	log "github.com/Archiit19/School-management/pkg/logger"
	pkgconfig "github.com/Archiit19/School-management/pkg/config"
	"github.com/Archiit19/School-management/exam-service/internal/config"
	"github.com/Archiit19/School-management/exam-service/internal/handler"
	"github.com/Archiit19/School-management/pkg/health"
	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/Archiit19/School-management/pkg/middleware"
	"github.com/Archiit19/School-management/pkg/server"
	"github.com/Archiit19/School-management/pkg/tracer"
	"github.com/Archiit19/School-management/pkg/userclient"
	"github.com/Archiit19/School-management/exam-service/internal/model"
	"github.com/Archiit19/School-management/exam-service/internal/repository"
	"github.com/Archiit19/School-management/exam-service/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Archiit19/School-management/exam-service/docs"
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
	if _, err := log.InitFromEnv("exam-service"); err != nil {
		log.Fatal("failed to initialize logger", log.Err(err))
	}

	traceShutdown, err := tracer.InitFromEnv("exam-service")
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

	if err := db.AutoMigrate(&model.Exam{}, &model.Mark{}); err != nil {
		log.Fatal("failed to migrate database", log.Err(err))
	}
	log.Info("database migrated")

	repo := repository.NewExamRepository(db)
	outbound := httpclient.OutboundHTTP("outbound")
	svc := service.NewExamService(repo, cfg, outbound)
	users := userclient.New(cfg.UserServiceURL, cfg.InternalServiceToken)
	h := handler.NewExamHandler(svc, users)

	r := middleware.NewEngine("exam-service")
	health.Register(r, "exam-service", health.CheckDB(db))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret,
		middleware.WithUUIDClaim("class_id"),
		middleware.WithUUIDClaim("section_id"),
	))
	{
		protected.GET("/exams/me", middleware.RequirePermission("view_own_exams"), h.GetMyExams)
		protected.GET("/exams", middleware.RequirePermission("view_exams"), h.GetExams)
		protected.POST("/exams", middleware.RequirePermission("create_exam"), h.CreateExam)
		protected.PATCH("/exams/:id", middleware.RequirePermission("create_exam"), h.UpdateExam)
		protected.POST("/exams/:id/complete", middleware.RequirePermission("create_exam"), h.CompleteExam)
		protected.DELETE("/exams/:id", middleware.RequirePermission("create_exam"), h.DeleteExam)
		protected.POST("/marks", middleware.RequirePermission("enter_marks"), h.EnterMarks)
		protected.POST("/results/publish", middleware.RequirePermission("publish_results"), h.PublishResults)
		protected.GET("/results/me", middleware.RequirePermission("view_own_results"), h.GetMyResults)
		protected.GET("/results", middleware.RequireAnyPermission("view_results", "publish_results", "enter_marks"), h.GetResults)
	}

	if err := server.Run(r, server.LoadConfigFromEnv(cfg.Port)); err != nil {
		log.Fatal("failed to start server", log.Err(err))
	}
}
