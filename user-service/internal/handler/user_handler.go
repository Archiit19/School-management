package handler

import (
	"net/http"

	"github.com/avaneeshravat/school-management/user-service/internal/model"
	"github.com/avaneeshravat/school-management/user-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// ─── Roles ──────────────────────────────────────────────────────────

// CreateRole godoc
// @Summary      Create a role
// @Description  Create a new role scoped to the authenticated user's school.
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateRoleRequest  true  "Role details"
// @Success      201   {object}  model.Role
// @Failure      400   {object}  model.ErrorResponse
// @Failure      401   {object}  model.ErrorResponse
// @Failure      409   {object}  model.ErrorResponse
// @Router       /api/v1/roles [post]
func (h *UserHandler) CreateRole(c *gin.Context) {
	var req model.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolID, _ := c.Get("school_id")

	role, err := h.svc.CreateRole(req, schoolID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// CreateRoleInternal godoc
// @Summary      Create a role (internal)
// @Description  Internal endpoint used by auth-service during school registration. School ID comes from the request body.
// @Tags         Internal
// @Accept       json
// @Produce      json
// @Param        body  body      model.CreateRoleRequest  true  "Role details with school_id"
// @Success      201   {object}  model.Role
// @Failure      400   {object}  model.ErrorResponse
// @Failure      409   {object}  model.ErrorResponse
// @Router       /api/v1/roles/internal [post]
func (h *UserHandler) CreateRoleInternal(c *gin.Context) {
	var req model.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.svc.CreateRoleInternal(req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// GetRoleByID godoc
// @Summary      Get role by ID
// @Description  Retrieve a single role by its UUID.
// @Tags         Roles
// @Produce      json
// @Param        id   path      string  true  "Role ID (UUID)"
// @Success      200  {object}  model.Role
// @Failure      400  {object}  model.ErrorResponse
// @Failure      404  {object}  model.ErrorResponse
// @Router       /api/v1/roles/{id} [get]
func (h *UserHandler) GetRoleByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	role, err := h.svc.GetRoleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	c.JSON(http.StatusOK, role)
}

// GetRoles godoc
// @Summary      List roles
// @Description  List all roles for the authenticated user's school.
// @Tags         Roles
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Role
// @Failure      401  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /api/v1/roles [get]
func (h *UserHandler) GetRoles(c *gin.Context) {
	schoolID, _ := c.Get("school_id")

	roles, err := h.svc.GetRolesBySchoolID(schoolID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, roles)
}

// ─── Permissions ────────────────────────────────────────────────────

// CreatePermission godoc
// @Summary      Create a permission
// @Description  Create a new system-level permission.
// @Tags         Permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreatePermissionRequest  true  "Permission details"
// @Success      201   {object}  model.Permission
// @Failure      400   {object}  model.ErrorResponse
// @Failure      401   {object}  model.ErrorResponse
// @Failure      409   {object}  model.ErrorResponse
// @Router       /api/v1/permissions [post]
func (h *UserHandler) CreatePermission(c *gin.Context) {
	var req model.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	perm, err := h.svc.CreatePermission(req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, perm)
}

// GetPermissions godoc
// @Summary      List permissions
// @Description  List all available permissions.
// @Tags         Permissions
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Permission
// @Failure      401  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /api/v1/permissions [get]
func (h *UserHandler) GetPermissions(c *gin.Context) {
	perms, err := h.svc.GetAllPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, perms)
}

// ─── Role-Permission Assignment ─────────────────────────────────────

// AssignPermission godoc
// @Summary      Assign permission to role
// @Description  Assign an existing permission to an existing role.
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.AssignPermissionRequest  true  "Role and permission IDs"
// @Success      201   {object}  model.RolePermission
// @Failure      400   {object}  model.ErrorResponse
// @Failure      401   {object}  model.ErrorResponse
// @Router       /api/v1/roles/assign-permission [post]
func (h *UserHandler) AssignPermission(c *gin.Context) {
	var req model.AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rp, err := h.svc.AssignPermissionToRole(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rp)
}

// GetRolePermissions godoc
// @Summary      Get role permissions
// @Description  List all permissions assigned to a specific role.
// @Tags         Roles
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Role ID (UUID)"
// @Success      200  {array}   model.Permission
// @Failure      400  {object}  model.ErrorResponse
// @Failure      401  {object}  model.ErrorResponse
// @Failure      500  {object}  model.ErrorResponse
// @Router       /api/v1/roles/{id}/permissions [get]
func (h *UserHandler) GetRolePermissions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	perms, err := h.svc.GetPermissionsByRoleID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, perms)
}
