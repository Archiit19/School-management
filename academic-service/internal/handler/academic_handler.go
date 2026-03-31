package handler

import (
	"net/http"

	"github.com/avaneeshravat/school-management/academic-service/internal/model"
	"github.com/avaneeshravat/school-management/academic-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AcademicHandler struct {
	svc *service.AcademicService
}

func NewAcademicHandler(svc *service.AcademicService) *AcademicHandler {
	return &AcademicHandler{svc: svc}
}

func (h *AcademicHandler) CreateClass(c *gin.Context) {
	var req model.CreateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	class, err := h.svc.CreateClass(req, schoolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, class)
}

func (h *AcademicHandler) CreateSection(c *gin.Context) {
	var req model.CreateSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	section, err := h.svc.CreateSection(req, schoolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, section)
}

func (h *AcademicHandler) CreateSubject(c *gin.Context) {
	var req model.CreateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	subject, err := h.svc.CreateSubject(req, schoolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, subject)
}

func (h *AcademicHandler) GetClasses(c *gin.Context) {
	schoolID := c.MustGet("school_id").(uuid.UUID)
	classes, err := h.svc.GetClasses(schoolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, classes)
}
