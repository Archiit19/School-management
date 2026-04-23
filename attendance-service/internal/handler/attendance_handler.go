package handler

import (
	"errors"
	"net/http"

	"github.com/avaneeshravat/school-management/attendance-service/internal/apierrors"
	"github.com/avaneeshravat/school-management/attendance-service/internal/model"
	"github.com/avaneeshravat/school-management/attendance-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func writeErr(c *gin.Context, err error) {
	var he *apierrors.HTTP
	if errors.As(err, &he) {
		c.JSON(he.Status, gin.H{"error": he.Error()})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

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
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, record)
}

// BulkCreateAttendance godoc
// @Summary      Bulk mark attendance
// @Description  Mark attendance for multiple students at once.
// @Tags         Attendance
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.BulkCreateAttendanceRequest  true  "Bulk payload"
// @Success      201   {object}  model.BulkAttendanceResponse
// @Failure      400   {object}  model.ErrorResponse
// @Router       /attendance/bulk [post]
func (h *AttendanceHandler) BulkCreateAttendance(c *gin.Context) {
	var req model.BulkCreateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	userID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	resp, err := h.svc.BulkCreateAttendance(req, schoolID, userID, roleName)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
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
		writeErr(c, err)
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
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, record)
}

func permissionsFromContext(c *gin.Context) []string {
	raw, ok := c.Get("permissions")
	if !ok {
		return nil
	}
	list, ok := raw.([]string)
	if !ok {
		return nil
	}
	return list
}

// CreateTeacherAttendance godoc
// @Summary      Mark teacher attendance
// @Description  Mark daily attendance for a teacher (self or another user with permission).
// @Tags         TeacherAttendance
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateTeacherAttendanceRequest  true  "Teacher attendance payload"
// @Success      201   {object}  model.TeacherAttendance
// @Failure      400   {object}  model.ErrorResponse
// @Router       /teacher-attendance [post]
func (h *AttendanceHandler) CreateTeacherAttendance(c *gin.Context) {
	var req model.CreateTeacherAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	userID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	record, err := h.svc.CreateTeacherAttendance(req, schoolID, userID, roleName, permissionsFromContext(c))
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, record)
}

// BulkCreateTeacherAttendance godoc
// @Summary      Bulk mark teacher attendance
// @Description  Mark attendance for multiple teachers (admin).
// @Tags         TeacherAttendance
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.BulkCreateTeacherAttendanceRequest  true  "Bulk payload"
// @Success      201   {object}  model.BulkTeacherAttendanceResponse
// @Failure      400   {object}  model.ErrorResponse
// @Router       /teacher-attendance/bulk [post]
func (h *AttendanceHandler) BulkCreateTeacherAttendance(c *gin.Context) {
	var req model.BulkCreateTeacherAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	userID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	resp, err := h.svc.BulkCreateTeacherAttendance(req, schoolID, userID, roleName, permissionsFromContext(c))
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetTeacherAttendance godoc
// @Summary      List teacher attendance
// @Description  View teacher attendance with filters and pagination.
// @Tags         TeacherAttendance
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  model.TeacherAttendanceListResponse
// @Failure      400  {object}  model.ErrorResponse
// @Router       /teacher-attendance [get]
func (h *AttendanceHandler) GetTeacherAttendance(c *gin.Context) {
	var query model.TeacherAttendanceQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	userID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	resp, err := h.svc.GetTeacherAttendance(schoolID, userID, roleName, permissionsFromContext(c), query)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateTeacherAttendance godoc
// @Summary      Edit teacher attendance
// @Description  Update status or remarks (recorder or super_admin).
// @Tags         TeacherAttendance
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true  "Teacher attendance ID"
// @Param        body  body      model.UpdateTeacherAttendanceRequest  true  "Update payload"
// @Success      200   {object}  model.TeacherAttendance
// @Failure      400   {object}  model.ErrorResponse
// @Router       /teacher-attendance/{id} [patch]
func (h *AttendanceHandler) UpdateTeacherAttendance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid teacher attendance id"})
		return
	}

	var req model.UpdateTeacherAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	userID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	record, err := h.svc.UpdateTeacherAttendance(id, req, schoolID, userID, roleName)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, record)
}
