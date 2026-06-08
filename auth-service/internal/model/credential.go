package model

import (
	"time"

	"github.com/google/uuid"
)

type UserCredential struct {
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	PasswordHash string    `json:"-" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRole struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:ux_user_school_role"`
	SchoolID  uuid.UUID `json:"school_id" gorm:"type:uuid;not null;uniqueIndex:ux_user_school_role"`
	RoleID    uuid.UUID `json:"role_id" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRoleMember struct {
	UserID   uuid.UUID `json:"user_id"`
	SchoolID uuid.UUID `json:"school_id"`
	RoleID   uuid.UUID `json:"role_id"`
}
