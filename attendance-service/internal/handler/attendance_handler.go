package handler

import (
	"net/http"

	"github.com/avaneeshravat/school-management/attendance-service/internal/model"
	"github.com/avaneeshravat/school-management/attendance-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AttendanceHandler struct {
	svc *service.AttendanceService
}

func NewAttendanceHandler(svc *service.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{svc: svc}
}

// CreateAttendance godoc
// @Summary      Mark attendance
// @Description  Teacher or super admin marks daily attendance for a student.
// @Tags         Attendance
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateAttendanceRequest  true  "Attendance payload"
// @Success      201   {object}  model.Attendance
// @Failure      400   {object}  model.ErrorResponse
// @Router       /attendance [post]
func (h *AttendanceHandler) CreateAttendance(c *gin.Context) {
	var req model.CreateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	userID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	record, err := h.svc.CreateAttendance(req, schoolID, userID, roleName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

// GetAttendance godoc
// @Summary      List attendance
// @Description  View attendance records with filters and pagination.
// @Tags         Attendance
// @Produce      json
// @Security     BearerAuth
// @Param        page       query     int     false  "Page"
// @Param        limit      query     int     false  "Limit"
// @Param        date       query     string  false  "Date (YYYY-MM-DD)"
// @Param        student_id query     string  false  "Student ID"
// @Param        class_id   query     string  false  "Class ID"
// @Param        section_id query     string  false  "Section ID"
// @Param        subject_id query     string  false  "Subject ID"
// @Param        status     query     string  false  "Attendance status"
// @Success      200        {object}  model.AttendanceListResponse
// @Failure      400        {object}  model.ErrorResponse
// @Router       /attendance [get]
func (h *AttendanceHandler) GetAttendance(c *gin.Context) {
	var query model.AttendanceQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	resp, err := h.svc.GetAttendance(schoolID, query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateAttendance godoc
// @Summary      Edit attendance
// @Description  Update attendance status or remarks for a record.
// @Tags         Attendance
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                      true  "Attendance ID"
// @Param        body  body      model.UpdateAttendanceRequest  true  "Update payload"
// @Success      200   {object}  model.Attendance
// @Failure      400   {object}  model.ErrorResponse
// @Router       /attendance/{id} [patch]
func (h *AttendanceHandler) UpdateAttendance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attendance id"})
		return
	}

	var req model.UpdateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	userID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	record, err := h.svc.UpdateAttendance(id, req, schoolID, userID, roleName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}
