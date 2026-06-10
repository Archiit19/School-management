package pupil

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChildValidator checks that a parent account is linked to a student user.
type ChildValidator interface {
	ParentHasChild(parentID, childID uuid.UUID) (bool, error)
}

// ResolveStudentID returns the pupil user ID for portal /me endpoints.
// Students use JWT student_id (or user_id). Parents must pass student_id query param.
func ResolveStudentID(c *gin.Context, validator ChildValidator) (uuid.UUID, error) {
	roleName, _ := c.Get("role_name")
	role, _ := roleName.(string)

	if strings.EqualFold(role, "student") {
		if sidVal, ok := c.Get("student_id"); ok {
			if sid, ok := sidVal.(uuid.UUID); ok && sid != uuid.Nil {
				return sid, nil
			}
		}
		uidVal, ok := c.Get("user_id")
		if !ok {
			return uuid.Nil, errors.New("user context missing")
		}
		uid, ok := uidVal.(uuid.UUID)
		if !ok || uid == uuid.Nil {
			return uuid.Nil, errors.New("invalid user context")
		}
		return uid, nil
	}

	if strings.EqualFold(role, "parent") {
		childStr := strings.TrimSpace(c.Query("student_id"))
		if childStr == "" {
			return uuid.Nil, errors.New("student_id query parameter is required for parent accounts")
		}
		childID, err := uuid.Parse(childStr)
		if err != nil {
			return uuid.Nil, errors.New("invalid student_id")
		}
		parentVal, ok := c.Get("user_id")
		if !ok {
			return uuid.Nil, errors.New("user context missing")
		}
		parentID, ok := parentVal.(uuid.UUID)
		if !ok || parentID == uuid.Nil {
			return uuid.Nil, errors.New("invalid user context")
		}
		if validator == nil {
			return uuid.Nil, errors.New("parent-child validation is not configured")
		}
		okChild, err := validator.ParentHasChild(parentID, childID)
		if err != nil {
			return uuid.Nil, err
		}
		if !okChild {
			return uuid.Nil, errors.New("you do not have access to this student")
		}
		return childID, nil
	}

	return uuid.Nil, errors.New("this account is not linked to a student record")
}

// WriteForbidden sends a 403 JSON error for pupil resolution failures.
func WriteForbidden(c *gin.Context, err error) {
	c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
}
