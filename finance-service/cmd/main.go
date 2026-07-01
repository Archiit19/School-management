package main

import (
	"context"
	"time"

	log "github.com/Archiit19/School-management/pkg/logger"
	pkgconfig "github.com/Archiit19/School-management/pkg/config"
	"github.com/Archiit19/School-management/pkg/health"
	"github.com/Archiit19/School-management/pkg/middleware"
	"github.com/Archiit19/School-management/pkg/server"
	"github.com/Archiit19/School-management/pkg/tracer"
	"github.com/Archiit19/School-management/finance-service/internal/config"
	"github.com/Archiit19/School-management/finance-service/internal/handler"
	"github.com/Archiit19/School-management/finance-service/internal/model"
	"github.com/Archiit19/School-management/finance-service/internal/repository"
	"github.com/Archiit19/School-management/finance-service/internal/service"
	"github.com/Archiit19/School-management/pkg/userclient"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	_ "github.com/Archiit19/School-management/finance-service/docs"
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
	if _, err := log.InitFromEnv("finance-service"); err != nil {
		log.Fatal("failed to initialize logger", log.Err(err))
	}

	traceShutdown, err := tracer.InitFromEnv("finance-service")
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

	db, err := pkgconfig.OpenGORM(cfg.DSN(), &pkgconfig.GORMOptions{
		Config: &gorm.Config{Logger: gormlogger.Discard},
	})
	if err != nil {
		log.Fatal("failed to connect to database", log.Err(err))
	}
	if err := tracer.InstrumentGORM(db); err != nil {
		log.Fatal("failed to instrument database", log.Err(err))
	}
	log.Info("connected to database")

	if err := db.AutoMigrate(&model.Fee{}, &model.Payment{}); err != nil {
		log.Fatal("failed to migrate database", log.Err(err))
	}
	log.Info("database migrated")

	repo := repository.NewFinanceRepository(db)
	svc := service.NewFinanceService(repo)
	users := userclient.New(cfg.UserServiceURL, cfg.InternalServiceToken)
	h := handler.NewFinanceHandler(svc, users)

	r := middleware.NewEngine("finance-service")
	health.Register(r, "finance-service", health.CheckDB(db))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.POST("/fees", middleware.RequirePermission("create_fee"), h.CreateFee)
		protected.POST("/payments", middleware.RequirePermission("record_payment"), h.RecordPayment)
		protected.GET("/dues/me", middleware.RequirePermission("view_own_dues"), h.GetMyDues)
		protected.GET("/dues", middleware.RequirePermission("view_dues"), h.GetDues)
	}

	if err := server.Run(r, server.LoadConfigFromEnv(cfg.Port)); err != nil {
		log.Fatal("failed to start server", log.Err(err))
	}
}
