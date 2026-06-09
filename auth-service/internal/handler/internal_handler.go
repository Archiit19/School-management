package handler

import (
	"net/http"

	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/Archiit19/School-management/auth-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InternalHandler struct {
	creds *service.CredentialService
	rbac  *service.RBACService
}

func NewInternalHandler(creds *service.CredentialService, rbac *service.RBACService) *InternalHandler {
	return &InternalHandler{creds: creds, rbac: rbac}
}

func (h *InternalHandler) SetCredential(c *gin.Context) {
	var req model.SetCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	if err := h.creds.SetPassword(userID, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "credential set"})
}

func (h *InternalHandler) DeleteCredential(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.creds.RemoveUserCompletely(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "credentials and roles removed"})
}

func (h *InternalHandler) AssignUserRole(c *gin.Context) {
	var req model.AssignUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := uuid.Parse(req.UserID)
	schoolID, _ := uuid.Parse(req.SchoolID)
	roleID, _ := uuid.Parse(req.RoleID)
	if err := h.creds.AssignUserRole(userID, schoolID, roleID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "role assigned"})
}

func (h *InternalHandler) UpdateUserRole(c *gin.Context) {
	var req model.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := uuid.Parse(req.UserID)
	schoolID, _ := uuid.Parse(req.SchoolID)
	roleID, _ := uuid.Parse(req.RoleID)
	if err := h.creds.UpdateUserRole(userID, schoolID, roleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

func (h *InternalHandler) RemoveUserRole(c *gin.Context) {
	var req model.RemoveUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := uuid.Parse(req.UserID)
	schoolID, _ := uuid.Parse(req.SchoolID)
	if err := h.creds.RemoveUserRole(userID, schoolID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role removed"})
}

func (h *InternalHandler) ListUserRoles(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if sid := c.Query("school_id"); sid != "" {
		schoolID, err := uuid.Parse(sid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school_id"})
			return
		}
		ur, err := h.creds.GetUserRole(userID, schoolID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
			return
		}
		roleName := h.rbac.RoleName(ur.RoleID)
		c.JSON(http.StatusOK, gin.H{
			"user_id":   userID,
			"school_id": schoolID,
			"role_id":   ur.RoleID,
			"role_name": roleName,
		})
		return
	}
	rows, err := h.creds.ListUserRoles(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]model.UserRoleMember, len(rows))
	for i, r := range rows {
		out[i] = model.UserRoleMember{UserID: r.UserID, SchoolID: r.SchoolID, RoleID: r.RoleID}
	}
	c.JSON(http.StatusOK, out)
}
