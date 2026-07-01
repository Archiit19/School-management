package handler

import (
	"net/http"

	log "github.com/Archiit19/School-management/pkg/logger"
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
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	schoolID, _ := c.Get("school_id")
	role, err := h.svc.CreateRole(req, schoolID.(uuid.UUID))
	if err != nil {
		logServiceError(c, http.StatusConflict, "create role failed", err, log.AddField("school_id", schoolID.(uuid.UUID)), log.AddField("name", req.Name))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("role created", log.AddField("role_id", role.ID), log.AddField("school_id", schoolID.(uuid.UUID)), log.AddField("name", role.Name))
	c.JSON(http.StatusCreated, role)
}

func (h *RBACHandler) CreateRoleInternal(c *gin.Context) {
	var req model.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role, err := h.svc.CreateRoleInternal(req)
	if err != nil {
		logServiceError(c, http.StatusConflict, "create role internal failed", err, log.AddField("name", req.Name))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("role created (internal)", log.AddField("role_id", role.ID), log.AddField("school_id", role.SchoolID), log.AddField("name", role.Name))
	c.JSON(http.StatusCreated, role)
}

func (h *RBACHandler) BootstrapSchoolInternal(c *gin.Context) {
	var req struct {
		SchoolID string `json:"school_id" binding:"required,uuid"`
	}
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
	superID, err := h.svc.BootstrapSchoolRoles(schoolID)
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "bootstrap school roles failed", err, log.AddField("school_id", schoolID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("school roles bootstrapped", log.AddField("school_id", schoolID), log.AddField("super_admin_role_id", superID))
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
		logServiceError(c, http.StatusNotFound, "get role by name internal failed", err, log.AddField("school_id", schoolID), log.AddField("name", name))
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
		logServiceError(c, http.StatusNotFound, "get role failed", err, log.AddField("role_id", id))
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *RBACHandler) GetRoles(c *gin.Context) {
	schoolID, _ := c.Get("school_id")
	roles, err := h.svc.GetRolesBySchoolID(schoolID.(uuid.UUID))
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "list roles failed", err, log.AddField("school_id", schoolID.(uuid.UUID)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *RBACHandler) GetPermissions(c *gin.Context) {
	perms, err := h.svc.GetAllPermissions()
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "list permissions failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, perms)
}

func (h *RBACHandler) AssignPermission(c *gin.Context) {
	var req model.AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rp, err := h.svc.AssignPermissionToRole(req)
	if err != nil {
		logServiceError(c, http.StatusBadRequest, "assign permission failed", err, log.AddField("role_id", req.RoleID), log.AddField("permission_id", req.PermissionID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("permission assigned to role", log.AddField("role_id", rp.RoleID), log.AddField("permission_id", rp.PermissionID))
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
		logServiceError(c, http.StatusInternalServerError, "get role permissions failed", err, log.AddField("role_id", id))
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
		logServiceError(c, http.StatusBadRequest, "remove permission from role failed", err, log.AddField("role_id", roleID), log.AddField("permission_id", permissionID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("permission removed from role", log.AddField("role_id", roleID), log.AddField("permission_id", permissionID))
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
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateRoleFields(id, req.Fields); err != nil {
		logServiceError(c, http.StatusBadRequest, "update role fields failed", err, log.AddField("role_id", id))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("role fields updated", log.AddField("role_id", id))
	c.JSON(http.StatusOK, gin.H{"fields": req.Fields})
}
