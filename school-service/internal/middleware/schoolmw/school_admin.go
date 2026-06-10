package schoolmw

import (
	"net/http"

	"github.com/Archiit19/School-management/school-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequireSchoolAdmin(svc *service.SchoolService) gin.HandlerFunc {
	return func(c *gin.Context) {
		schoolID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid school id"})
			return
		}

		userID := c.MustGet("user_id").(uuid.UUID)
		ok, err := svc.IsUserMember(schoolID, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "not an admin of this school"})
			return
		}
		c.Next()
	}
}
