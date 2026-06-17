package main

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/Archiit19/School-management/pkg/logger"
	"github.com/Archiit19/School-management/exam-service/internal/config"
	"github.com/Archiit19/School-management/exam-service/internal/handler"
	"github.com/Archiit19/School-management/exam-service/internal/model"
	"github.com/Archiit19/School-management/exam-service/internal/repository"
	"github.com/Archiit19/School-management/exam-service/internal/service"
	"github.com/Archiit19/School-management/pkg/middleware"
	"github.com/Archiit19/School-management/pkg/userclient"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

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

	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Discard,
	})
	if err != nil {
		log.Fatal("failed to connect to database", log.Err(err))
	}
	log.Info("connected to database")

	if err := db.AutoMigrate(&model.Exam{}, &model.Mark{}); err != nil {
		log.Fatal("failed to migrate database", log.Err(err))
	}
	log.Info("database migrated")

	repo := repository.NewExamRepository(db)
	httpClient := &http.Client{Timeout: 8 * time.Second}
	svc := service.NewExamService(repo, cfg, httpClient)
	users := userclient.New(cfg.UserServiceURL, cfg.InternalServiceToken)
	h := handler.NewExamHandler(svc, users)

	r := middleware.NewEngine()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "exam-service is running"})
	})
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
		protected.POST("/marks", middleware.RequirePermission("enter_marks"), h.EnterMarks)
		protected.POST("/results/publish", middleware.RequirePermission("publish_results"), h.PublishResults)
		protected.GET("/results/me", middleware.RequirePermission("view_own_results"), h.GetMyResults)
		protected.GET("/results", middleware.RequirePermission("view_results"), h.GetResults)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Info("starting http server",
		log.AddField("addr", addr),
		log.AddField("swagger", fmt.Sprintf("http://localhost%s/swagger/index.html", addr)),
	)
	if err := r.Run(addr); err != nil {
		log.Fatal("failed to start server", log.Err(err))
	}
}
