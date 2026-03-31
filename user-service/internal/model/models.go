package model

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a role scoped to a school.
type Role struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()" example:"550e8400-e29b-41d4-a716-446655440000"`
	SchoolID    uuid.UUID `json:"school_id" gorm:"type:uuid;not null;index" example:"550e8400-e29b-41d4-a716-446655440001"`
	Name        string    `json:"name" gorm:"not null" example:"teacher"`
	Description string    `json:"description" example:"Teacher role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a system-level permission.
type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()" example:"550e8400-e29b-41d4-a716-446655440010"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null" example:"manage_students"`
	Description string    `json:"description" example:"Can manage student records"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RolePermission is the join table between roles and permissions.
type RolePermission struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()" example:"550e8400-e29b-41d4-a716-446655440020"`
	RoleID       uuid.UUID `json:"role_id" gorm:"type:uuid;not null;index" example:"550e8400-e29b-41d4-a716-446655440000"`
	PermissionID uuid.UUID `json:"permission_id" gorm:"type:uuid;not null;index" example:"550e8400-e29b-41d4-a716-446655440010"`
	CreatedAt    time.Time `json:"created_at"`
}

// ─── Request / Response DTOs ────────────────────────────────────────

// CreateRoleRequest is the payload for creating a new role.
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required" example:"teacher"`
	Description string `json:"description" example:"Teacher role"`
	SchoolID    string `json:"school_id" example:"550e8400-e29b-41d4-a716-446655440001"` // set from JWT for public API, from body for internal
}

// CreatePermissionRequest is the payload for creating a new permission.
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required" example:"manage_students"`
	Description string `json:"description" example:"Can manage student records"`
}

// AssignPermissionRequest is the payload for assigning a permission to a role.
type AssignPermissionRequest struct {
	RoleID       string `json:"role_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	PermissionID string `json:"permission_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440010"`
}

// ErrorResponse represents a generic error response.
type ErrorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}
