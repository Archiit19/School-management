package handler

import (
	"net/http"

	"github.com/avaneeshravat/school-management/auth-service/internal/model"
	"github.com/avaneeshravat/school-management/auth-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// RegisterSchool godoc
// @Summary      Register a new school
// @Description  Creates a new school, a super_admin role, and the first admin user. Returns a JWT token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.RegisterSchoolRequest  true  "School and admin details"
// @Success      201   {object}  model.RegisterSchoolResponse
// @Failure      400   {object}  model.ErrorResponse
// @Failure      409   {object}  model.ErrorResponse
// @Router       /auth/register-school [post]
func (h *AuthHandler) RegisterSchool(c *gin.Context) {
	var req model.RegisterSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.RegisterSchool(req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary      Login
// @Description  Authenticate with email and password. Returns a JWT token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      model.LoginRequest  true  "Login credentials"
// @Success      200   {object}  model.LoginResponse
// @Failure      400   {object}  model.ErrorResponse
// @Failure      401   {object}  model.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMe godoc
// @Summary      Get current user
// @Description  Returns the profile of the currently authenticated user.
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  model.User
// @Failure      401  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Router       /auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	user, err := h.svc.GetMe(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if perms, ok := c.Get("permissions"); ok {
		if permList, ok := perms.([]string); ok {
			user.Permissions = permList
		}
	}

	c.JSON(http.StatusOK, user)
}
