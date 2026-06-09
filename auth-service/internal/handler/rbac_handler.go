package handler

import (
	"net/http"

	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/Archiit19/School-management/auth-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RBACHandler struct {
	svc *service.RBACService
}

func NewRBACHandler(svc *service.RBACService) *RBACHandler {
	return &RBACHandler{svc: svc}
}

func (h *RBACHandler) CreateRole(c *gin.Context) {
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

func (h *RBACHandler) CreateRoleInternal(c *gin.Context) {
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

func (h *RBACHandler) BootstrapSchoolInternal(c *gin.Context) {
	var req struct {
		SchoolID string `json:"school_id" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	schoolID, err := uuid.Parse(req.SchoolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school_id"})
		return
	}
	superID, err := h.svc.BootstrapSchoolRoles(schoolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"super_admin_role_id": superID.String()})
}

func (h *RBACHandler) GetRoleByNameAndSchoolInternal(c *gin.Context) {
	schoolStr := c.Query("school_id")
	name := c.Query("name")
	if schoolStr == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "school_id and name query params are required"})
		return
	}
	schoolID, err := uuid.Parse(schoolStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school_id"})
		return
	}
	role, err := h.svc.GetRoleByNameAndSchool(name, schoolID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *RBACHandler) GetRoleByID(c *gin.Context) {
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

func (h *RBACHandler) GetRoles(c *gin.Context) {
	schoolID, _ := c.Get("school_id")
	roles, err := h.svc.GetRolesBySchoolID(schoolID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *RBACHandler) GetPermissions(c *gin.Context) {
	perms, err := h.svc.GetAllPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, perms)
}

func (h *RBACHandler) AssignPermission(c *gin.Context) {
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

func (h *RBACHandler) GetRolePermissions(c *gin.Context) {
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

func (h *RBACHandler) RemovePermissionFromRole(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
	permissionID, err := uuid.Parse(c.Param("permissionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission id"})
		return
	}
	if err := h.svc.RemovePermissionFromRole(roleID, permissionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "permission removed from role"})
}

func (h *RBACHandler) GetRoleFields(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
	fields, err := h.svc.GetRoleFields(id)
	if err != nil {
		c.JSON(http.StatusOK, []model.FieldDefinition{})
		return
	}
	c.JSON(http.StatusOK, fields)
}

func (h *RBACHandler) UpdateRoleFields(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}
	var req model.UpdateRoleFieldsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateRoleFields(id, req.Fields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"fields": req.Fields})
}
