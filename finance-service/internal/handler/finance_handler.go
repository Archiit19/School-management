package handler

import (
	"net/http"

	"github.com/avaneeshravat/school-management/finance-service/internal/model"
	"github.com/avaneeshravat/school-management/finance-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FinanceHandler struct {
	svc *service.FinanceService
}

func NewFinanceHandler(svc *service.FinanceService) *FinanceHandler {
	return &FinanceHandler{svc: svc}
}

// CreateFee godoc
// @Summary      Create fee
// @Description  Create fee structure and optionally assign it to class/section/student.
// @Tags         Finance
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateFeeRequest  true  "Fee payload"
// @Success      201   {object}  model.Fee
// @Failure      400   {object}  model.ErrorResponse
// @Router       /fees [post]
func (h *FinanceHandler) CreateFee(c *gin.Context) {
	var req model.CreateFeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	userID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	fee, err := h.svc.CreateFee(req, schoolID, userID, roleName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, fee)
}

// RecordPayment godoc
// @Summary      Record payment
// @Description  Record a payment against assigned fee.
// @Tags         Finance
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreatePaymentRequest  true  "Payment payload"
// @Success      201   {object}  model.Payment
// @Failure      400   {object}  model.ErrorResponse
// @Router       /payments [post]
func (h *FinanceHandler) RecordPayment(c *gin.Context) {
	var req model.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	userID := c.MustGet("user_id").(uuid.UUID)
	roleName := c.MustGet("role_name").(string)

	payment, err := h.svc.RecordPayment(req, schoolID, userID, roleName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, payment)
}

// GetDues godoc
// @Summary      Get dues
// @Description  Get outstanding dues by student/class/section filters.
// @Tags         Finance
// @Produce      json
// @Security     BearerAuth
// @Param        student_id  query     string  false  "Student ID"
// @Param        class_id    query     string  false  "Class ID"
// @Param        section_id  query     string  false  "Section ID"
// @Success      200         {array}   model.DueItem
// @Failure      400         {object}  model.ErrorResponse
// @Router       /dues [get]
func (h *FinanceHandler) GetDues(c *gin.Context) {
	var query model.DueQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID := c.MustGet("school_id").(uuid.UUID)
	dues, err := h.svc.GetDues(schoolID, query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dues)
}

// GetMyDues godoc
// @Summary      My dues (pupil portal)
// @Description  Outstanding dues for the JWT student_id only — student_id query param is ignored.
// @Tags         Finance
// @Produce      json
// @Security     BearerAuth
// @Param        class_id    query     string  false  "Class ID"
// @Param        section_id  query     string  false  "Section ID"
// @Success      200         {array}   model.DueItem
// @Failure      400         {object}  model.ErrorResponse
// @Failure      403         {object}  model.ErrorResponse
// @Router       /dues/me [get]
func (h *FinanceHandler) GetMyDues(c *gin.Context) {
	sidVal, ok := c.Get("student_id")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "this account is not linked to a student record"})
		return
	}
	studentID := sidVal.(uuid.UUID)

	var query model.DueQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query.StudentID = studentID.String()

	schoolID := c.MustGet("school_id").(uuid.UUID)
	dues, err := h.svc.GetDues(schoolID, query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dues)
}
