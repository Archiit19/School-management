package model

import (
	"time"

	"github.com/google/uuid"
)

type School struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"not null"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserSchool maps an auth user to a school with a school-scoped role.
type UserSchool struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index;uniqueIndex:ux_user_school"`
	SchoolID  uuid.UUID `json:"school_id" gorm:"type:uuid;not null;index;uniqueIndex:ux_user_school"`
	RoleID    uuid.UUID `json:"role_id" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserSchoolMember struct {
	UserID   uuid.UUID `json:"user_id"`
	SchoolID uuid.UUID `json:"school_id"`
	RoleID   uuid.UUID `json:"role_id"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	RoleID string `json:"role_id" binding:"required,uuid"`
}

type UpdateMemberRequest struct {
	RoleID string `json:"role_id" binding:"required,uuid"`
}

type CreateSchoolWithAdminRequest struct {
	UserID  string `json:"user_id" binding:"required,uuid"`
	Name    string `json:"name" binding:"required"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email" binding:"required,email"`
}

type CreateSchoolRequest struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email" binding:"required,email"`
}

type UpdateSchoolRequest struct {
	Name     *string `json:"name"`
	Address  *string `json:"address"`
	Phone    *string `json:"phone"`
	Email    *string `json:"email"`
	IsActive *bool   `json:"is_active"`
}

type SchoolListQuery struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
	Search   string `form:"search"`
	IsActive *bool  `form:"is_active"`
}

type SchoolListResponse struct {
	Schools []School `json:"schools"`
	Total   int64    `json:"total"`
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
