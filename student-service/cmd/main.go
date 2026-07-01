package main

import (
	"context"
	"time"

	log "github.com/Archiit19/School-management/pkg/logger"
	pkgconfig "github.com/Archiit19/School-management/pkg/config"
	"github.com/Archiit19/School-management/student-service/internal/config"
	"github.com/Archiit19/School-management/student-service/internal/handler"
	"github.com/Archiit19/School-management/pkg/health"
	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/Archiit19/School-management/pkg/middleware"
	"github.com/Archiit19/School-management/pkg/server"
	"github.com/Archiit19/School-management/pkg/tracer"
	"github.com/Archiit19/School-management/student-service/internal/model"
	"github.com/Archiit19/School-management/student-service/internal/repository"
	"github.com/Archiit19/School-management/student-service/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Archiit19/School-management/student-service/docs"
)

// @title           Student Service API
// @version         1.0
// @description     Student admission and management service.
// @host            localhost:8084
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	if _, err := log.InitFromEnv("student-service"); err != nil {
		log.Fatal("failed to initialize logger", log.Err(err))
	}

	traceShutdown, err := tracer.InitFromEnv("student-service")
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

	if err := db.AutoMigrate(&model.Student{}); err != nil {
		log.Fatal("failed to migrate database", log.Err(err))
	}
	log.Info("database migrated")

	repo := repository.NewStudentRepository(db)
	httpCfg := pkgconfig.LoadHTTPClientConfigFromEnv()
	userInternal := httpclient.NewFromConfig(httpclient.ClientConfig{
		BaseURL: cfg.UserServiceURL,
		Token:   cfg.InternalServiceToken,
		Name:    "user-service",
		HTTP:    &httpCfg,
	})
	outbound := httpclient.OutboundHTTP("outbound")
	svc := service.NewStudentService(repo, cfg, userInternal, outbound)
	h := handler.NewStudentHandler(svc)

	r := middleware.NewEngine("student-service")
	health.Register(r, "student-service", health.CheckDB(db))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/internal/students/:id", h.GetStudentByIDInternal)

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.POST("/students", middleware.RequirePermission("admit_student"), h.CreateStudent)
		protected.GET("/students/me", middleware.RequirePermission("view_own_profile"), h.GetMyStudentRecord)
		protected.GET("/students", middleware.RequirePermission("view_students"), h.GetStudents)
		protected.PATCH("/students/:id", middleware.RequirePermission("update_student"), h.UpdateStudent)
	}

	if err := server.Run(r, server.LoadConfigFromEnv(cfg.Port)); err != nil {
		log.Fatal("failed to start server", log.Err(err))
	}
}
