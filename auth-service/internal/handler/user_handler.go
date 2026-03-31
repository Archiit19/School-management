package handler

import (
	"net/http"

	"github.com/avaneeshravat/school-management/auth-service/internal/model"
	"github.com/avaneeshravat/school-management/auth-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	svc *service.UserManagementService
}

func NewUserHandler(svc *service.UserManagementService) *UserHandler {
	return &UserHandler{svc: svc}
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Admin creates a new user (teacher, staff, parent). Requires super_admin role.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateUserRequest  true  "User details"
// @Success      201   {object}  model.User
// @Failure      400   {object}  model.ErrorResponse
// @Failure      401   {object}  model.ErrorResponse
// @Failure      403   {object}  model.ErrorResponse
// @Failure      409   {object}  model.ErrorResponse
// @Router       /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID, _ := c.Get("school_id")

	user, err := h.svc.CreateUser(req, schoolID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUsers godoc
// @Summary      List users
// @Description  List all users for the school with optional filtering and pagination. Requires super_admin role.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int     false  "Page number"      default(1)
// @Param        limit     query     int     false  "Items per page"   default(20)
// @Param        search    query     string  false  "Search by name or email"
// @Param        role_id   query     string  false  "Filter by role ID"
// @Param        is_active query     bool    false  "Filter by active status"
// @Success      200       {object}  model.UserListResponse
// @Failure      400       {object}  model.ErrorResponse
// @Failure      401       {object}  model.ErrorResponse
// @Failure      403       {object}  model.ErrorResponse
// @Router       /users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	var query model.UserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID, _ := c.Get("school_id")

	resp, err := h.svc.GetUsers(schoolID.(uuid.UUID), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Get a single user by their ID (scoped to the admin's school). Requires super_admin role.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  model.User
// @Failure      400  {object}  model.ErrorResponse
// @Failure      401  {object}  model.ErrorResponse
// @Failure      403  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Router       /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	schoolID, _ := c.Get("school_id")

	user, err := h.svc.GetUserByID(id, schoolID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser godoc
// @Summary      Update a user
// @Description  Partially update a user's name, email, role, or active status. Requires super_admin role.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                  true  "User ID (UUID)"
// @Param        body  body      model.UpdateUserRequest  true  "Fields to update"
// @Success      200   {object}  model.User
// @Failure      400   {object}  model.ErrorResponse
// @Failure      401   {object}  model.ErrorResponse
// @Failure      403   {object}  model.ErrorResponse
// @Router       /users/{id} [patch]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID, _ := c.Get("school_id")

	user, err := h.svc.UpdateUser(id, req, schoolID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Permanently delete a user. Cannot delete yourself. Requires super_admin role.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  model.MessageResponse
// @Failure      400  {object}  model.ErrorResponse
// @Failure      401  {object}  model.ErrorResponse
// @Failure      403  {object}  model.ErrorResponse
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	schoolID, _ := c.Get("school_id")
	requestingUserID, _ := c.Get("user_id")

	err = h.svc.DeleteUser(id, schoolID.(uuid.UUID), requestingUserID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}
