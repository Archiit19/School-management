package handler

import (
	"net/http"

	"github.com/avaneeshravat/school-management/exam-service/internal/model"
	"github.com/avaneeshravat/school-management/exam-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExamHandler struct {
	svc *service.ExamService
}

func NewExamHandler(svc *service.ExamService) *ExamHandler {
	return &ExamHandler{svc: svc}
}

// CreateExam godoc
// @Summary      Create exam
// @Description  Create a new exam for class/subject.
// @Tags         Exams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateExamRequest  true  "Exam payload"
// @Success      201   {object}  model.Exam
// @Failure      400   {object}  model.ErrorResponse
// @Router       /exams [post]
func (h *ExamHandler) CreateExam(c *gin.Context) {
	var req model.CreateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	userID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	exam, err := h.svc.CreateExam(req, schoolID, userID, roleName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, exam)
}

// EnterMarks godoc
// @Summary      Enter marks
// @Description  Enter or update marks for a student in an exam.
// @Tags         Exams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.EnterMarksRequest  true  "Marks payload"
// @Success      201   {object}  model.Mark
// @Failure      400   {object}  model.ErrorResponse
// @Router       /marks [post]
func (h *ExamHandler) EnterMarks(c *gin.Context) {
	var req model.EnterMarksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	userID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	mark, err := h.svc.EnterMarks(req, schoolID, userID, roleName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, mark)
}

// PublishResults godoc
// @Summary      Publish results
// @Description  Publish exam results for report card visibility.
// @Tags         Results
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.PublishResultRequest  true  "Publish payload"
// @Success      200   {object}  model.Exam
// @Failure      400   {object}  model.ErrorResponse
// @Router       /results/publish [post]
func (h *ExamHandler) PublishResults(c *gin.Context) {
	var req model.PublishResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	exam, err := h.svc.PublishResults(req, schoolID, roleName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, exam)
}

// GetResults godoc
// @Summary      Get results
// @Description  Fetch computed result entries for report cards.
// @Tags         Results
// @Produce      json
// @Security     BearerAuth
// @Param        exam_id     query     string  false  "Exam ID"
// @Param        student_id  query     string  false  "Student ID"
// @Param        class_id    query     string  false  "Class ID"
// @Success      200         {array}   model.ResultItem
// @Failure      400         {object}  model.ErrorResponse
// @Router       /results [get]
func (h *ExamHandler) GetResults(c *gin.Context) {
	var query model.ResultQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)
	results, err := h.svc.GetResults(schoolID, query, roleName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// GetMyResults godoc
// @Summary      My exam results (pupil portal)
// @Description  Returns published results for the JWT student_id only — student_id query param is ignored.
// @Tags         Results
// @Produce      json
// @Security     BearerAuth
// @Param        exam_id   query     string  false  "Exam ID"
// @Param        class_id  query     string  false  "Class ID"
// @Success      200       {array}   model.ResultItem
// @Failure      400       {object}  model.ErrorResponse
// @Failure      403       {object}  model.ErrorResponse
// @Router       /results/me [get]
func (h *ExamHandler) GetMyResults(c *gin.Context) {
	sidVal, ok := c.Get("student_id")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "this account is not linked to a student record"})
		return
	}
	studentID := sidVal.(uuid.UUID)

	var query model.ResultQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query.StudentID = studentID.String()

	schoolID := c.MustGet("school_id").(uuid.UUID)
	results, err := h.svc.GetResults(schoolID, query, "student")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}
