package handler

import (
	"net/http"

	log "github.com/Archiit19/School-management/pkg/logger"
	"github.com/Archiit19/School-management/exam-service/internal/model"
	"github.com/Archiit19/School-management/exam-service/internal/service"
	"github.com/Archiit19/School-management/pkg/pupil"
	"github.com/Archiit19/School-management/pkg/userclient"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ExamHandler struct {
	svc   *service.ExamService
	users *userclient.Client
}

func NewExamHandler(svc *service.ExamService, users *userclient.Client) *ExamHandler {
	return &ExamHandler{svc: svc, users: users}
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
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	userID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	exam, err := h.svc.CreateExam(req, schoolID, userID, roleName)
	if err != nil {
		logServiceError(c, http.StatusBadRequest, "create exam failed", err, log.AddField("school_id", schoolID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("exam created", log.AddField("exam_id", exam.ID), log.AddField("school_id", schoolID), log.AddField("title", exam.Title))
	c.JSON(http.StatusCreated, exam)
}

// UpdateExam godoc
// @Summary      Update exam
// @Description  Update exam details (not allowed after results are published).
// @Tags         Exams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true  "Exam ID"
// @Param        body  body      model.UpdateExamRequest  true  "Update payload"
// @Success      200   {object}  model.Exam
// @Failure      400   {object}  model.ErrorResponse
// @Router       /exams/{id} [patch]
func (h *ExamHandler) UpdateExam(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	var req model.UpdateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	exam, err := h.svc.UpdateExam(examID, req, schoolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, exam)
}

// CompleteExam godoc
// @Summary      Mark exam complete
// @Description  Mark an exam as conducted/complete (not allowed after results are published).
// @Tags         Exams
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Exam ID"
// @Success      200  {object}  model.Exam
// @Failure      400  {object}  model.ErrorResponse
// @Router       /exams/{id}/complete [post]
func (h *ExamHandler) CompleteExam(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	exam, err := h.svc.CompleteExam(examID, schoolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, exam)
}

// DeleteExam godoc
// @Summary      Delete exam
// @Description  Delete an exam and its marks (not allowed after results are published).
// @Tags         Exams
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Exam ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  model.ErrorResponse
// @Router       /exams/{id} [delete]
func (h *ExamHandler) DeleteExam(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	if err := h.svc.DeleteExam(examID, schoolID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "exam deleted"})
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
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	userID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	mark, err := h.svc.EnterMarks(req, schoolID, userID, roleName)
	if err != nil {
		logServiceError(c, http.StatusBadRequest, "enter marks failed", err, log.AddField("school_id", schoolID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("marks entered", log.AddField("mark_id", mark.ID), log.AddField("school_id", schoolID), log.AddField("exam_id", mark.ExamID), log.AddField("student_id", mark.StudentID))
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
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	exam, err := h.svc.PublishResults(req, schoolID, roleName)
	if err != nil {
		logServiceError(c, http.StatusBadRequest, "publish results failed", err, log.AddField("school_id", schoolID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("results published", log.AddField("exam_id", exam.ID), log.AddField("school_id", schoolID))
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
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)
	results, err := h.svc.GetResults(schoolID, query, roleName, permissionsFromContext(c))
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "get results failed", err, log.AddField("school_id", schoolID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// GetExams godoc
// @Summary      List exams
// @Description  List exams for the school. Supports filtering by class, section, subject, upcoming and published.
// @Tags         Exams
// @Produce      json
// @Security     BearerAuth
// @Param        class_id    query     string  false  "Class ID (UUID)"
// @Param        section_id  query     string  false  "Section ID (UUID)"
// @Param        subject_id  query     string  false  "Subject ID (UUID)"
// @Param        upcoming    query     bool    false  "Only exams scheduled for today or later"
// @Param        published   query     bool    false  "Filter by publish status"
// @Success      200         {array}   model.Exam
// @Failure      400         {object}  model.ErrorResponse
// @Router       /exams [get]
func (h *ExamHandler) GetExams(c *gin.Context) {
	var query model.ExamQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	schoolID := c.MustGet("school_id").(uuid.UUID)

	exams, err := h.svc.GetExams(schoolID, query)
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "list exams failed", err, log.AddField("school_id", schoolID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, exams)
}

// GetMyExams godoc
// @Summary      My class's exam schedule (pupil portal)
// @Description  Lists exams scheduled for the class linked to the JWT student_id.
// @Tags         Exams
// @Produce      json
// @Security     BearerAuth
// @Param        upcoming  query     bool  false  "Only exams scheduled for today or later (defaults to true)"
// @Success      200       {array}   model.Exam
// @Failure      403       {object}  model.ErrorResponse
// @Router       /exams/me [get]
func (h *ExamHandler) GetMyExams(c *gin.Context) {
	studentID, err := pupil.ResolveStudentID(c, h.users)
	if err != nil {
		userID, _ := c.Get("user_id")
		logServiceError(c, http.StatusForbidden, "get my exams failed", err, log.AddField("user_id", userID.(uuid.UUID)))
		pupil.WriteForbidden(c, err)
		return
	}
	schoolID := c.MustGet("school_id").(uuid.UUID)
	authHeader := c.GetHeader("Authorization")

	upcoming := true
	if v := c.Query("upcoming"); v == "false" || v == "0" {
		upcoming = false
	}

	exams, err := h.svc.GetMyExams(c.Request.Context(), schoolID, studentID, authHeader, upcoming)
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "get my exams failed", err, log.AddField("school_id", schoolID), log.AddField("student_id", studentID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, exams)
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
	studentID, err := pupil.ResolveStudentID(c, h.users)
	if err != nil {
		userID, _ := c.Get("user_id")
		logServiceError(c, http.StatusForbidden, "get my results failed", err, log.AddField("user_id", userID.(uuid.UUID)))
		pupil.WriteForbidden(c, err)
		return
	}

	var query model.ResultQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query.StudentID = studentID.String()

	schoolID := c.MustGet("school_id").(uuid.UUID)
	results, err := h.svc.GetResults(schoolID, query, "student", nil)
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "get my results failed", err, log.AddField("school_id", schoolID), log.AddField("student_id", studentID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
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
