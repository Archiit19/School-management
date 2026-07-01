package handler

import (
	"net/http"

	log "github.com/Archiit19/School-management/pkg/logger"
	"github.com/Archiit19/School-management/academic-service/internal/model"
	"github.com/Archiit19/School-management/academic-service/internal/service"
	"github.com/Archiit19/School-management/pkg/pupil"
	"github.com/Archiit19/School-management/pkg/userclient"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EnrollmentHandler struct {
	svc   *service.AcademicService
	users *userclient.Client
}

func NewEnrollmentHandler(svc *service.AcademicService, users *userclient.Client) *EnrollmentHandler {
	return &EnrollmentHandler{svc: svc, users: users}
}

func (h *EnrollmentHandler) ListEnrollments(c *gin.Context) {
	var query model.EnrollmentQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	schoolID := c.MustGet("school_id").(uuid.UUID)
	resp, err := h.svc.ListEnrollments(schoolID, query)
	if err != nil {
		logServiceError(c, http.StatusBadRequest, "list enrollments failed", err, log.AddField("school_id", schoolID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EnrollmentHandler) GetMyEnrollment(c *gin.Context) {
	schoolID := c.MustGet("school_id").(uuid.UUID)
	studentID, err := pupil.ResolveStudentID(c, h.users)
	if err != nil {
		logServiceError(c, http.StatusForbidden, "resolve student for my enrollment failed", err)
		pupil.WriteForbidden(c, err)
		return
	}
	enrollment, err := h.svc.GetEnrollmentByUser(studentID, schoolID)
	if err != nil {
		logServiceError(c, http.StatusNotFound, "get my enrollment failed", err, log.AddField("school_id", schoolID), log.AddField("student_id", studentID))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, enrollment)
}

func (h *EnrollmentHandler) UpsertEnrollmentInternal(c *gin.Context) {
	var req model.UpsertEnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	enrollment, err := h.svc.UpsertEnrollment(req)
	if err != nil {
		logServiceError(c, http.StatusBadRequest, "upsert enrollment internal failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("enrollment upserted", log.AddField("user_id", enrollment.UserID), log.AddField("school_id", enrollment.SchoolID), log.AddField("class_id", enrollment.ClassID))
	c.JSON(http.StatusOK, enrollment)
}

func (h *EnrollmentHandler) GetEnrollmentInternal(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	schoolStr := c.Query("school_id")
	if schoolStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "school_id query param required"})
		return
	}
	schoolID, err := uuid.Parse(schoolStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school_id"})
		return
	}
	enrollment, err := h.svc.GetEnrollmentByUser(userID, schoolID)
	if err != nil {
		logServiceError(c, http.StatusNotFound, "get enrollment internal failed", err, log.AddField("user_id", userID), log.AddField("school_id", schoolID))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, enrollment)
}

func (h *EnrollmentHandler) DeleteEnrollmentInternal(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	schoolStr := c.Query("school_id")
	if schoolStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "school_id query param required"})
		return
	}
	schoolID, err := uuid.Parse(schoolStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school_id"})
		return
	}
	if err := h.svc.DeleteEnrollment(userID, schoolID); err != nil {
		logServiceError(c, http.StatusBadRequest, "delete enrollment internal failed", err, log.AddField("user_id", userID), log.AddField("school_id", schoolID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("enrollment deleted", log.AddField("user_id", userID), log.AddField("school_id", schoolID))
	c.JSON(http.StatusOK, gin.H{"message": "enrollment deleted"})
}
