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

// CreateClass godoc
// @Summary      Create class
// @Description  Create a class for the authenticated school.
// @Tags         Academic
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateClassRequest  true  "Class payload"
// @Success      201   {object}  model.Class
// @Failure      400   {object}  model.ErrorResponse
// @Router       /classes [post]
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

// CreateSection godoc
// @Summary      Create section
// @Description  Create a section under a class.
// @Tags         Academic
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateSectionRequest  true  "Section payload"
// @Success      201   {object}  model.Section
// @Failure      400   {object}  model.ErrorResponse
// @Router       /sections [post]
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

// CreateSubject godoc
// @Summary      Create subject
// @Description  Create a subject for a class and optional section.
// @Tags         Academic
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateSubjectRequest  true  "Subject payload"
// @Success      201   {object}  model.Subject
// @Failure      400   {object}  model.ErrorResponse
// @Router       /subjects [post]
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

// GetClasses godoc
// @Summary      Get classes
// @Description  List classes with sections and subjects for the authenticated school.
// @Tags         Academic
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.ClassWithChildren
// @Failure      500  {object}  model.ErrorResponse
// @Router       /classes [get]
func (h *AcademicHandler) GetClasses(c *gin.Context) {
	schoolID := c.MustGet("school_id").(uuid.UUID)
	classes, err := h.svc.GetClasses(schoolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, classes)
}

// CreateTeacherAssignment godoc
// @Summary      Assign teacher
// @Description  Assign a teacher user to class and subject.
// @Tags         TeacherAssignments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateTeacherAssignmentRequest  true  "Assignment payload"
// @Success      201   {object}  model.TeacherAssignment
// @Failure      400   {object}  model.ErrorResponse
// @Router       /teacher-assignments [post]
func (h *AcademicHandler) CreateTeacherAssignment(c *gin.Context) {
	var req model.CreateTeacherAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	authHeader := c.GetHeader("Authorization")
	assignment, err := h.svc.CreateTeacherAssignment(req, schoolID, authHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, assignment)
}

// GetTeacherAssignments godoc
// @Summary      List teacher assignments
// @Description  List teacher assignments with optional filters.
// @Tags         TeacherAssignments
// @Produce      json
// @Security     BearerAuth
// @Param        teacher_user_id  query     string  false  "Teacher User ID"
// @Param        class_id         query     string  false  "Class ID"
// @Param        subject_id       query     string  false  "Subject ID"
// @Success      200              {array}   model.TeacherAssignment
// @Failure      500              {object}  model.ErrorResponse
// @Router       /teacher-assignments [get]
func (h *AcademicHandler) GetTeacherAssignments(c *gin.Context) {
	var query model.TeacherAssignmentQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	assignments, err := h.svc.GetTeacherAssignments(schoolID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, assignments)
}

// CreateAssignment godoc
// @Summary      Create assignment
// @Description  Teacher or super admin creates assignment and material.
// @Tags         Assignments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateAssignmentRequest  true  "Assignment payload"
// @Success      201   {object}  model.Assignment
// @Failure      400   {object}  model.ErrorResponse
// @Router       /assignments [post]
func (h *AcademicHandler) CreateAssignment(c *gin.Context) {
	var req model.CreateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	requestingUserID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	assignment, err := h.svc.CreateAssignment(req, schoolID, requestingUserID, roleName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, assignment)
}

// GetAssignments godoc
// @Summary      List assignments
// @Description  List assignments with optional filters.
// @Tags         Assignments
// @Produce      json
// @Security     BearerAuth
// @Param        class_id    query     string  false  "Class ID"
// @Param        subject_id  query     string  false  "Subject ID"
// @Param        teacher_id  query     string  false  "Teacher User ID"
// @Success      200         {array}   model.Assignment
// @Failure      500         {object}  model.ErrorResponse
// @Router       /assignments [get]
func (h *AcademicHandler) GetAssignments(c *gin.Context) {
	var query model.AssignmentQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	assignments, err := h.svc.GetAssignments(schoolID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, assignments)
}

// GetMyAssignments godoc
// @Summary      My class's assignments (pupil portal)
// @Description  Lists assignments for the class linked to the JWT student_id.
// @Tags         Assignments
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Assignment
// @Failure      400  {object}  model.ErrorResponse
// @Failure      403  {object}  model.ErrorResponse
// @Router       /assignments/me [get]
func (h *AcademicHandler) GetMyAssignments(c *gin.Context) {
	sidVal, ok := c.Get("student_id")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "this account is not linked to a student record"})
		return
	}
	studentID := sidVal.(uuid.UUID)
	schoolID := c.MustGet("school_id").(uuid.UUID)
	authHeader := c.GetHeader("Authorization")

	assignments, err := h.svc.GetMyAssignments(schoolID, studentID, authHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assignments)
}

// GetMySubmissions godoc
// @Summary      My submissions (pupil portal)
// @Description  Lists submissions where student_id matches JWT.
// @Tags         Assignments
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Submission
// @Failure      403  {object}  model.ErrorResponse
// @Router       /submissions/me [get]
func (h *AcademicHandler) GetMySubmissions(c *gin.Context) {
	sidVal, ok := c.Get("student_id")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "this account is not linked to a student record"})
		return
	}
	studentID := sidVal.(uuid.UUID)
	schoolID := c.MustGet("school_id").(uuid.UUID)

	subs, err := h.svc.GetMySubmissions(schoolID, studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subs)
}

// CreateMySubmission godoc
// @Summary      Submit my assignment (pupil portal)
// @Description  Pupil submits work; student_id is taken from JWT (no student_id in body).
// @Tags         Assignments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateMySubmissionRequest  true  "Submission payload"
// @Success      201   {object}  model.Submission
// @Failure      400   {object}  model.ErrorResponse
// @Failure      403   {object}  model.ErrorResponse
// @Router       /submissions/me [post]
func (h *AcademicHandler) CreateMySubmission(c *gin.Context) {
	sidVal, ok := c.Get("student_id")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "this account is not linked to a student record"})
		return
	}
	studentID := sidVal.(uuid.UUID)

	var req model.CreateMySubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	submission, err := h.svc.CreateOwnSubmission(req, schoolID, studentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, submission)
}

// CreateSubmission godoc
// @Summary      Create submission
// @Description  Submit student work for an assignment.
// @Tags         Assignments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateSubmissionRequest  true  "Submission payload"
// @Success      201   {object}  model.Submission
// @Failure      400   {object}  model.ErrorResponse
// @Router       /submissions [post]
func (h *AcademicHandler) CreateSubmission(c *gin.Context) {
	var req model.CreateSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	submittedBy := c.MustGet("user_id").(uuid.UUID)
	submission, err := h.svc.CreateSubmission(req, schoolID, submittedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, submission)
}
