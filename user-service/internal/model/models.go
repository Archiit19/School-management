package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	SchoolID *uuid.UUID             `json:"school_id,omitempty" gorm:"-"`
	RoleID   *uuid.UUID             `json:"role_id,omitempty" gorm:"-"`
	RoleName string                 `json:"role_name,omitempty" gorm:"-"`
	RoleData map[string]interface{} `json:"role_data,omitempty" gorm:"-"`
}

type CreateUserRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Email    string                 `json:"email" binding:"required,email"`
	Password string                 `json:"password" binding:"required,min=6"`
	RoleID   string                 `json:"role_id" binding:"required,uuid"`
	RoleData map[string]interface{} `json:"role_data"`
}

type UpdateUserRequest struct {
	Name     *string                `json:"name"`
	Email    *string                `json:"email"`
	RoleID   *string                `json:"role_id"`
	IsActive *bool                  `json:"is_active"`
	RoleData map[string]interface{} `json:"role_data"`
}

type UserListQuery struct {
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`
	Search    string `form:"search"`
	RoleID    string `form:"role_id"`
	IDs       string `form:"ids"`
	IsActive  *bool  `form:"is_active"`
	ClassID   string `form:"class_id"`
	SectionID string `form:"section_id"`
}

type UserListResponse struct {
	Users []User `json:"users"`
	Total int64  `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

type CreateProfileInternalRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
