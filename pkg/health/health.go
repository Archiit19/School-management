package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Checker validates a dependency. A non-nil error marks the service as not ready.
type Checker func(ctx context.Context) error

// Register mounts liveness and readiness endpoints on r.
// /health and /health/ready run readiness checks; /health/live is always OK when the process is up.
func Register(r *gin.Engine, service string, checks ...Checker) {
	r.GET("/health/live", liveness(service))
	ready := readiness(service, checks)
	r.GET("/health/ready", ready)
	r.GET("/health", ready)
}

func liveness(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": service,
			"check":   "live",
		})
	}
}

func readiness(service string, checks []Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		for _, check := range checks {
			if check == nil {
				continue
			}
			if err := check(ctx); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status":  "unavailable",
					"service": service,
					"check":   "ready",
					"error":   err.Error(),
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": service,
			"check":   "ready",
		})
	}
}

// CheckDB verifies PostgreSQL connectivity.
func CheckDB(db *gorm.DB) Checker {
	return func(ctx context.Context) error {
		if db == nil {
			return errMissing("database")
		}
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	}
}

type errMissing string

func (e errMissing) Error() string {
	return string(e) + " is not configured"
}
