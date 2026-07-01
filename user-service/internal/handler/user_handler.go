package handler

import (
	"net/http"

	log "github.com/Archiit19/School-management/pkg/logger"
	"github.com/Archiit19/School-management/user-service/internal/model"
	"github.com/Archiit19/School-management/user-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	schoolID, _ := c.Get("school_id")
	user, err := h.svc.CreateUser(c.Request.Context(), req, schoolID.(uuid.UUID))
	if err != nil {
		logServiceError(c, http.StatusConflict, "create user failed", err, log.AddField("school_id", schoolID.(uuid.UUID)))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("user created", log.AddField("user_id", user.ID), log.AddField("school_id", schoolID.(uuid.UUID)), log.AddField("email", user.Email))
	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	var query model.UserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	schoolID, _ := c.Get("school_id")
	resp, err := h.svc.GetUsers(c.Request.Context(), schoolID.(uuid.UUID), query)
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "list users failed", err, log.AddField("school_id", schoolID.(uuid.UUID)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	schoolID, _ := c.Get("school_id")
	user, err := h.svc.GetUserByID(c.Request.Context(), id, schoolID.(uuid.UUID))
	if err != nil {
		logServiceError(c, http.StatusNotFound, "get user failed", err, log.AddField("user_id", id), log.AddField("school_id", schoolID.(uuid.UUID)))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetUserMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	schoolID, _ := c.Get("school_id")
	user, err := h.svc.GetUserMe(c.Request.Context(), userID.(uuid.UUID), schoolID.(uuid.UUID))
	if err != nil {
		logServiceError(c, http.StatusNotFound, "get own profile failed", err, log.AddField("user_id", userID.(uuid.UUID)))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetMyChildren(c *gin.Context) {
	userID, _ := c.Get("user_id")
	schoolID, _ := c.Get("school_id")
	children, err := h.svc.GetMyChildren(c.Request.Context(), userID.(uuid.UUID), schoolID.(uuid.UUID))
	if err != nil {
		logServiceError(c, http.StatusForbidden, "list children failed", err, log.AddField("user_id", userID.(uuid.UUID)))
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"children": children})
}

func (h *UserHandler) GetChildForParent(c *gin.Context) {
	parentID, _ := c.Get("user_id")
	schoolID, _ := c.Get("school_id")
	childID, err := uuid.Parse(c.Param("childId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid child id"})
		return
	}
	child, err := h.svc.GetChildForParent(c.Request.Context(), parentID.(uuid.UUID), childID, schoolID.(uuid.UUID))
	if err != nil {
		logServiceError(c, http.StatusNotFound, "get child for parent failed", err,
			log.AddField("user_id", parentID.(uuid.UUID)), log.AddField("child_id", childID))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, child)
}

func (h *UserHandler) ParentHasChildInternal(c *gin.Context) {
	parentID, err := uuid.Parse(c.Param("parentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent id"})
		return
	}
	childID, err := uuid.Parse(c.Param("childId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid child id"})
		return
	}
	ok, err := h.svc.ParentHasChild(c.Request.Context(), parentID, childID)
	if err != nil {
		logServiceError(c, http.StatusInternalServerError, "parent has-child check failed", err,
			log.AddField("parent_id", parentID), log.AddField("child_id", childID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "child not linked to parent"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"linked": true})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	schoolID, _ := c.Get("school_id")
	user, err := h.svc.UpdateUser(c.Request.Context(), id, req, schoolID.(uuid.UUID))
	if err != nil {
		logServiceError(c, http.StatusBadRequest, "update user failed", err, log.AddField("user_id", id), log.AddField("school_id", schoolID.(uuid.UUID)))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("user updated", log.AddField("user_id", id), log.AddField("school_id", schoolID.(uuid.UUID)))
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	schoolID, _ := c.Get("school_id")
	requestingUserID, _ := c.Get("user_id")
	if err := h.svc.DeleteUser(c.Request.Context(), id, schoolID.(uuid.UUID), requestingUserID.(uuid.UUID)); err != nil {
		logServiceError(c, http.StatusBadRequest, "delete user failed", err, log.AddField("user_id", id), log.AddField("school_id", schoolID.(uuid.UUID)))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("user deleted", log.AddField("user_id", id), log.AddField("school_id", schoolID.(uuid.UUID)))
	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

func (h *UserHandler) CreateProfileInternal(c *gin.Context) {
	var req model.CreateProfileInternalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.svc.CreateProfileInternal(c.Request.Context(), req)
	if err != nil {
		logServiceError(c, http.StatusConflict, "create profile internal failed", err, log.AddField("email", req.Email))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("profile created (internal)", log.AddField("user_id", user.ID), log.AddField("email", user.Email))
	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUserInternal(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var schoolID *uuid.UUID
	if sid := c.Query("school_id"); sid != "" {
		parsed, err := uuid.Parse(sid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid school_id"})
			return
		}
		schoolID = &parsed
	}
	user, err := h.svc.GetUserForInternal(c.Request.Context(), id, schoolID)
	if err != nil {
		logServiceError(c, http.StatusNotFound, "get user internal failed", err, log.AddField("user_id", id))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetUserProfileInternal(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	profile, err := h.svc.GetUserProfileInternal(c.Request.Context(), id)
	if err != nil {
		logServiceError(c, http.StatusNotFound, "get user profile internal failed", err, log.AddField("user_id", id))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *UserHandler) GetUserByEmailInternal(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email query param required"})
		return
	}
	user, err := h.svc.GetUserForInternalByEmail(c.Request.Context(), email)
	if err != nil {
		logServiceError(c, http.StatusNotFound, "get user by email internal failed", err, log.AddField("email", email))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateProfileInternal(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req struct {
		Name  *string `json:"name"`
		Email *string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logBindError(c, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.svc.UpdateProfileInternal(c.Request.Context(), id, req.Name, req.Email)
	if err != nil {
		logServiceError(c, http.StatusConflict, "update profile internal failed", err, log.AddField("user_id", id))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("profile updated (internal)", log.AddField("user_id", id))
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteProfileInternal(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.svc.DeleteProfileInternal(c.Request.Context(), id); err != nil {
		logServiceError(c, http.StatusBadRequest, "delete profile internal failed", err, log.AddField("user_id", id))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	requestLogger(c).Info("profile deleted (internal)", log.AddField("user_id", id))
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}
