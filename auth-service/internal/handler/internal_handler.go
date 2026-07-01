package handler

import (
	"net/http"

	log "github.com/Archiit19/School-management/pkg/logger"
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
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	if err := h.creds.SetPassword(userID, req.Password); err != nil {
		logServiceError(c, http.StatusInternalServerError, "set credential failed", err, log.AddField("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("credential set", log.AddField("user_id", userID))
	c.JSON(http.StatusOK, gin.H{"message": "credential set"})
}

func (h *InternalHandler) DeleteCredential(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.creds.RemoveUserCompletely(userID); err != nil {
		logServiceError(c, http.StatusInternalServerError, "delete credential failed", err, log.AddField("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("credentials and roles removed", log.AddField("user_id", userID))
	c.JSON(http.StatusOK, gin.H{"message": "credentials and roles removed"})
}

func (h *InternalHandler) AssignUserRole(c *gin.Context) {
	var req model.AssignUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := uuid.Parse(req.UserID)
	schoolID, _ := uuid.Parse(req.SchoolID)
	roleID, _ := uuid.Parse(req.RoleID)
	if err := h.creds.AssignUserRole(userID, schoolID, roleID); err != nil {
		logServiceError(c, http.StatusConflict, "assign user role failed", err, log.AddField("user_id", userID), log.AddField("school_id", schoolID), log.AddField("role_id", roleID))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("user role assigned", log.AddField("user_id", userID), log.AddField("school_id", schoolID), log.AddField("role_id", roleID))
	c.JSON(http.StatusCreated, gin.H{"message": "role assigned"})
}

func (h *InternalHandler) UpdateUserRole(c *gin.Context) {
	var req model.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := uuid.Parse(req.UserID)
	schoolID, _ := uuid.Parse(req.SchoolID)
	roleID, _ := uuid.Parse(req.RoleID)
	if err := h.creds.UpdateUserRole(userID, schoolID, roleID); err != nil {
		logServiceError(c, http.StatusBadRequest, "update user role failed", err, log.AddField("user_id", userID), log.AddField("school_id", schoolID), log.AddField("role_id", roleID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("user role updated", log.AddField("user_id", userID), log.AddField("school_id", schoolID), log.AddField("role_id", roleID))
	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

func (h *InternalHandler) RemoveUserRole(c *gin.Context) {
	var req model.RemoveUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := uuid.Parse(req.UserID)
	schoolID, _ := uuid.Parse(req.SchoolID)
	if err := h.creds.RemoveUserRole(userID, schoolID); err != nil {
		logServiceError(c, http.StatusBadRequest, "remove user role failed", err, log.AddField("user_id", userID), log.AddField("school_id", schoolID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("user role removed", log.AddField("user_id", userID), log.AddField("school_id", schoolID))
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
			logServiceError(c, http.StatusNotFound, "get user role failed", err, log.AddField("user_id", userID), log.AddField("school_id", schoolID))
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
		logServiceError(c, http.StatusInternalServerError, "list user roles failed", err, log.AddField("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]model.UserRoleMember, len(rows))
	for i, r := range rows {
		out[i] = model.UserRoleMember{UserID: r.UserID, SchoolID: r.SchoolID, RoleID: r.RoleID}
	}
	c.JSON(http.StatusOK, out)
}
