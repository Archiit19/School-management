package handler

import (
	"net/http"

	"github.com/avaneeshravat/school-management/student-service/internal/model"
	"github.com/avaneeshravat/school-management/student-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StudentHandler struct {
	svc *service.StudentService
}

func NewStudentHandler(svc *service.StudentService) *StudentHandler {
	return &StudentHandler{svc: svc}
}

// CreateStudent godoc
// @Summary      Create student
// @Description  Create student, assign class/section, and optionally link parent.
// @Tags         Students
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateStudentRequest  true  "Student payload"
// @Success      201   {object}  model.Student
// @Failure      400   {object}  model.ErrorResponse
// @Router       /students [post]
func (h *StudentHandler) CreateStudent(c *gin.Context) {
	var req model.CreateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	authHeader := c.GetHeader("Authorization")

	student, err := h.svc.CreateStudent(req, schoolID, authHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, student)
}

// GetStudents godoc
// @Summary      List students
// @Description  List students for authenticated school with filters and pagination.
// @Tags         Students
// @Produce      json
// @Security     BearerAuth
// @Param        page            query     int     false  "Page"
// @Param        limit           query     int     false  "Limit"
// @Param        search          query     string  false  "Search by first/last name"
// @Param        class_id        query     string  false  "Class ID"
// @Param        section_id      query     string  false  "Section ID"
// @Param        parent_user_id  query     string  false  "Parent User ID"
// @Param        is_active       query     bool    false  "Active status"
// @Success      200             {object}  model.StudentListResponse
// @Failure      400             {object}  model.ErrorResponse
// @Failure      500             {object}  model.ErrorResponse
// @Router       /students [get]
func (h *StudentHandler) GetStudents(c *gin.Context) {
	var query model.StudentListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	resp, err := h.svc.GetStudents(schoolID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateStudent godoc
// @Summary      Update student
// @Description  Update student details, class/section assignment, and parent link.
// @Tags         Students
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                   true  "Student ID"
// @Param        body  body      model.UpdateStudentRequest  true  "Update payload"
// @Success      200   {object}  model.Student
// @Failure      400   {object}  model.ErrorResponse
// @Router       /students/{id} [patch]
func (h *StudentHandler) UpdateStudent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}

	var req model.UpdateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	authHeader := c.GetHeader("Authorization")

	student, err := h.svc.UpdateStudent(id, req, schoolID, authHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, student)
}
