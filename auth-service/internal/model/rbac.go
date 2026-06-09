package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// FieldDefinition describes one input the UI should collect when creating a user with this role.
type FieldDefinition struct {
	Key      string   `json:"key"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Options  []string `json:"options,omitempty"`
}

// RoleField stores the per-role profile field schema (one row per role).
type RoleField struct {
	RoleID    uuid.UUID       `json:"role_id" gorm:"type:uuid;primaryKey"`
	Fields    json.RawMessage `json:"fields" gorm:"type:jsonb;not null;default:'[]'"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type Role struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID    uuid.UUID `json:"school_id" gorm:"type:uuid;not null;index"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RolePermission struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RoleID       uuid.UUID `json:"role_id" gorm:"type:uuid;not null;index"`
	PermissionID uuid.UUID `json:"permission_id" gorm:"type:uuid;not null;index"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateRoleRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	SchoolID    string            `json:"school_id"`
	Fields      []FieldDefinition `json:"fields"`
}

type UpdateRoleFieldsRequest struct {
	Fields []FieldDefinition `json:"fields" binding:"required"`
}

type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type AssignPermissionRequest struct {
	RoleID       string `json:"role_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
}

type SetCredentialRequest struct {
	UserID   string `json:"user_id" binding:"required,uuid"`
	Password string `json:"password" binding:"required,min=6"`
}

type AssignUserRoleRequest struct {
	UserID   string `json:"user_id" binding:"required,uuid"`
	SchoolID string `json:"school_id" binding:"required,uuid"`
	RoleID   string `json:"role_id" binding:"required,uuid"`
}

type UpdateUserRoleRequest struct {
	UserID   string `json:"user_id" binding:"required,uuid"`
	SchoolID string `json:"school_id" binding:"required,uuid"`
	RoleID   string `json:"role_id" binding:"required,uuid"`
}

type RemoveUserRoleRequest struct {
	UserID   string `json:"user_id" binding:"required,uuid"`
	SchoolID string `json:"school_id" binding:"required,uuid"`
}
