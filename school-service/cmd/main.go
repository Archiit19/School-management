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
	"github.com/Archiit19/School-management/school-service/internal/config"
	"github.com/Archiit19/School-management/school-service/internal/handler"
	"github.com/Archiit19/School-management/school-service/internal/migrate"
	"github.com/Archiit19/School-management/school-service/internal/middleware/schoolmw"
	"github.com/Archiit19/School-management/school-service/internal/model"
	"github.com/Archiit19/School-management/school-service/internal/repository"
	"github.com/Archiit19/School-management/school-service/internal/service"
)

// @title           School Service API
// @version         1.0
// @description     School registry with admin mapping and CRUD.
// @host            localhost:8088
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	if _, err := log.InitFromEnv("school-service"); err != nil {
		log.Fatal("failed to initialize logger", log.Err(err))
	}

	traceShutdown, err := tracer.InitFromEnv("school-service")
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

	if err := db.AutoMigrate(&model.School{}, &model.UserSchool{}); err != nil {
		log.Fatal("failed to migrate database", log.Err(err))
	}
	if err := migrate.DropRoleIDFromMemberships(db); err != nil {
		log.Fatal("failed membership schema migration", log.Err(err))
	}
	log.Info("database migrated")

	repo := repository.NewSchoolRepository(db)
	svc := service.NewSchoolService(repo, cfg.AuthServiceURL, cfg.InternalServiceToken)
	h := handler.NewSchoolHandler(svc)

	r := middleware.NewEngine("school-service")
	health.Register(r, "school-service", health.CheckDB(db))

	internal := r.Group("/internal")
	internal.Use(middleware.RequireInternalToken(cfg.InternalServiceToken))
	{
		internal.POST("/schools/with-admin", h.CreateSchoolWithAdminInternal)
		internal.GET("/schools/by-email", h.GetSchoolByEmailInternal)
		internal.GET("/schools/by-user/:userId", h.ListSchoolsByUserInternal)
		internal.GET("/users/:userId/memberships", h.ListMembershipsForUserInternal)
		internal.GET("/schools/:id/members", h.ListMembersForSchoolInternal)
		internal.POST("/schools/:id/members", h.AddMemberInternal)
		internal.GET("/schools/:id/members/:userId", h.GetMemberInternal)
		internal.DELETE("/schools/:id/members/:userId", h.RemoveMemberInternal)
		internal.GET("/schools/:id/admins/:userId", h.CheckAdminInternal)
		internal.GET("/schools/:id", h.GetSchoolInternal)
	}

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.GET("/schools/mine", middleware.RequirePermission("view_my_schools"), h.ListMySchools)
		protected.POST("/schools", middleware.RequirePermission("create_school"), h.CreateSchool)
		protected.GET("/schools", middleware.RequirePermission("view_all_schools"), h.ListSchools)
		protected.GET("/schools/me", middleware.RequirePermission("view_school"), h.GetMySchool)
		protected.PATCH("/schools/me", middleware.RequirePermission("manage_school"), h.UpdateMySchool)

		adminSchool := protected.Group("/schools/:id")
		adminSchool.Use(schoolmw.RequireSchoolAdmin(svc))
		{
			adminSchool.GET("", middleware.RequirePermission("view_school"), h.GetSchool)
			adminSchool.PATCH("", middleware.RequirePermission("manage_school"), h.UpdateSchool)
			adminSchool.DELETE("", middleware.RequirePermission("manage_school"), h.DeleteSchool)
		}
	}

	if err := server.Run(r, server.LoadConfigFromEnv(cfg.Port)); err != nil {
		log.Fatal("failed to start server", log.Err(err))
	}
}
