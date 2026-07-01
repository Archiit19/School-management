package handler

import (
	"net/http"

	log "github.com/Archiit19/School-management/pkg/logger"
	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/Archiit19/School-management/auth-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req model.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Signup(c.Request.Context(), req)
	if err != nil {
		logServiceError(c, http.StatusConflict, "signup failed", err, log.AddField("email", req.Email))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("user signed up", log.AddField("user_id", resp.User.ID), log.AddField("email", resp.User.Email))
	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) RegisterSchool(c *gin.Context) {
	var req model.RegisterSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.RegisterSchool(c.Request.Context(), req)
	if err != nil {
		logServiceError(c, http.StatusConflict, "register school failed", err, log.AddField("school_email", req.SchoolEmail))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("school registered",
		log.AddField("school_id", resp.School.ID),
		log.AddField("admin_user_id", resp.Admin.ID),
		log.AddField("school_name", resp.School.Name),
	)
	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		logServiceError(c, http.StatusUnauthorized, "login failed", err, log.AddField("email", req.Email))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) SelectSchool(c *gin.Context) {
	var req model.SelectSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID, err := uuid.Parse(req.SchoolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school_id"})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	resp, err := h.svc.SelectSchool(c.Request.Context(), userID, schoolID)
	if err != nil {
		logServiceError(c, http.StatusForbidden, "select school failed", err, log.AddField("user_id", userID), log.AddField("school_id", schoolID))
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("school selected", log.AddField("user_id", userID), log.AddField("school_id", schoolID))
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) ExitSchool(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	resp, err := h.svc.ExitSchool(c.Request.Context(), userID)
	if err != nil {
		logServiceError(c, http.StatusBadRequest, "exit school failed", err, log.AddField("user_id", userID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("exited school context", log.AddField("user_id", userID))
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	user, err := h.svc.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		logServiceError(c, http.StatusConflict, "update profile failed", err, log.AddField("user_id", userID))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("profile updated", log.AddField("user_id", userID))
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	schoolID, _ := c.Get("school_id")
	sid, _ := schoolID.(uuid.UUID)

	var jwtPerms []string
	if perms, ok := c.Get("permissions"); ok {
		if permList, ok := perms.([]string); ok {
			jwtPerms = permList
		}
	}

	resp, err := h.svc.GetMe(c.Request.Context(), userID, sid, jwtPerms)
	if err != nil {
		logServiceError(c, http.StatusNotFound, "get me failed", err, log.AddField("user_id", userID))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
