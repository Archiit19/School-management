package handler

import (
	"net/http"

	log "github.com/Archiit19/School-management/pkg/logger"
	"github.com/Archiit19/School-management/school-service/internal/model"
	"github.com/Archiit19/School-management/school-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SchoolHandler struct {
	svc *service.SchoolService
}

func NewSchoolHandler(svc *service.SchoolService) *SchoolHandler {
	return &SchoolHandler{svc: svc}
}

func (h *SchoolHandler) CreateSchool(c *gin.Context) {
	var req model.CreateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	school, err := h.svc.CreateSchoolForUser(userID, req)
	if err != nil {
		logServiceError(c, http.StatusConflict, "create school failed", err, log.AddField("user_id", userID))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("school created", log.AddField("school_id", school.ID), log.AddField("user_id", userID), log.AddField("name", school.Name))
	c.JSON(http.StatusCreated, school)
}

func (h *SchoolHandler) ListMySchools(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	schools, err := h.svc.ListSchoolsForUser(userID)
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "list my schools failed", err, log.AddField("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if schools == nil {
		schools = []model.School{}
	}
	c.JSON(http.StatusOK, gin.H{"schools": schools, "total": len(schools)})
}

func (h *SchoolHandler) ListSchools(c *gin.Context) {
	var query model.SchoolListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.ListSchools(query)
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "list schools failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *SchoolHandler) GetSchool(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school id"})
		return
	}

	school, err := h.svc.GetSchool(id)
	if err != nil {
		logServiceError(c, http.StatusNotFound, "get school failed", err, log.AddField("school_id", id))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, school)
}

func (h *SchoolHandler) GetMySchool(c *gin.Context) {
	schoolID := c.MustGet("school_id").(uuid.UUID)
	if schoolID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no school selected — use select-school first"})
		return
	}
	school, err := h.svc.GetSchool(schoolID)
	if err != nil {
		logServiceError(c, http.StatusNotFound, "get my school failed", err, log.AddField("school_id", schoolID))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, school)
}

func (h *SchoolHandler) UpdateSchool(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school id"})
		return
	}

	var req model.UpdateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	school, err := h.svc.UpdateSchool(id, req)
	if err != nil {
		logServiceError(c, http.StatusConflict, "update school failed", err, log.AddField("school_id", id))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("school updated", log.AddField("school_id", school.ID))
	c.JSON(http.StatusOK, school)
}

func (h *SchoolHandler) UpdateMySchool(c *gin.Context) {
	schoolID := c.MustGet("school_id").(uuid.UUID)
	if schoolID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no school selected"})
		return
	}

	var req model.UpdateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	school, err := h.svc.UpdateSchool(schoolID, req)
	if err != nil {
		logServiceError(c, http.StatusConflict, "update my school failed", err, log.AddField("school_id", schoolID))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("school updated", log.AddField("school_id", school.ID))
	c.JSON(http.StatusOK, school)
}

func (h *SchoolHandler) DeleteSchool(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school id"})
		return
	}

	if err := h.svc.DeleteSchool(id); err != nil {
		logServiceError(c, http.StatusNotFound, "delete school failed", err, log.AddField("school_id", id))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("school deactivated", log.AddField("school_id", id))
	c.JSON(http.StatusOK, gin.H{"message": "school deactivated"})
}

func (h *SchoolHandler) CreateSchoolWithAdminInternal(c *gin.Context) {
	var req model.CreateSchoolWithAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	school, err := h.svc.CreateSchoolForUser(userID, model.CreateSchoolRequest{
		Name:    req.Name,
		Address: req.Address,
		Phone:   req.Phone,
		Email:   req.Email,
	})
	if err != nil {
		logServiceError(c, http.StatusConflict, "create school with admin internal failed", err, log.AddField("user_id", userID))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("school created (internal)", log.AddField("school_id", school.ID), log.AddField("user_id", userID), log.AddField("name", school.Name))
	c.JSON(http.StatusCreated, school)
}

func (h *SchoolHandler) GetSchoolInternal(c *gin.Context) {
	h.GetSchool(c)
}

func (h *SchoolHandler) GetSchoolByEmailInternal(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email query param is required"})
		return
	}

	school, err := h.svc.GetSchoolByEmail(email)
	if err != nil {
		logServiceError(c, http.StatusNotFound, "get school by email internal failed", err, log.AddField("email", email))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, school)
}

func (h *SchoolHandler) ListSchoolsByUserInternal(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	schools, err := h.svc.ListSchoolsForUser(userID)
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "list schools by user internal failed", err, log.AddField("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if schools == nil {
		schools = []model.School{}
	}
	c.JSON(http.StatusOK, schools)
}

func (h *SchoolHandler) CheckAdminInternal(c *gin.Context) {
	schoolID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school id"})
		return
	}
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	ok, err := h.svc.IsUserMember(schoolID, userID)
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "check admin internal failed", err, log.AddField("school_id", schoolID), log.AddField("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not a member"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *SchoolHandler) ListMembershipsForUserInternal(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	rows, err := h.svc.ListMembershipsForUser(userID)
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "list memberships for user internal failed", err, log.AddField("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []model.UserSchoolMember{}
	}
	c.JSON(http.StatusOK, rows)
}

func (h *SchoolHandler) ListMembersForSchoolInternal(c *gin.Context) {
	schoolID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school id"})
		return
	}
	rows, err := h.svc.ListMembersForSchool(schoolID)
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "list members for school internal failed", err, log.AddField("school_id", schoolID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []model.UserSchoolMember{}
	}
	c.JSON(http.StatusOK, rows)
}

func (h *SchoolHandler) GetMemberInternal(c *gin.Context) {
	schoolID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school id"})
		return
	}
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	m, err := h.svc.GetMembership(schoolID, userID)
	if err != nil {
		logServiceError(c, http.StatusNotFound, "get member internal failed", err, log.AddField("school_id", schoolID), log.AddField("user_id", userID))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.UserSchoolMember{
		UserID:   m.UserID,
		SchoolID: m.SchoolID,
	})
}

func (h *SchoolHandler) AddMemberInternal(c *gin.Context) {
	schoolID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school id"})
		return
	}
	var req model.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	m, err := h.svc.AddMember(schoolID, userID)
	if err != nil {
		logServiceError(c, http.StatusConflict, "add member internal failed", err, log.AddField("school_id", schoolID), log.AddField("user_id", userID))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("school member added", log.AddField("school_id", m.SchoolID), log.AddField("user_id", m.UserID))
	c.JSON(http.StatusCreated, model.UserSchoolMember{
		UserID:   m.UserID,
		SchoolID: m.SchoolID,
	})
}

func (h *SchoolHandler) RemoveMemberInternal(c *gin.Context) {
	schoolID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school id"})
		return
	}
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.svc.RemoveMember(schoolID, userID); err != nil {
		logServiceError(c, http.StatusNotFound, "remove member internal failed", err, log.AddField("school_id", schoolID), log.AddField("user_id", userID))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("school member removed", log.AddField("school_id", schoolID), log.AddField("user_id", userID))
	c.JSON(http.StatusOK, gin.H{"message": "membership removed"})
}
