package model

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID     uuid.UUID  `json:"school_id" gorm:"type:uuid;not null;index"`
	FirstName    string     `json:"first_name" gorm:"not null"`
	LastName     string     `json:"last_name" gorm:"not null"`
	ParentUserID *uuid.UUID `json:"parent_user_id,omitempty" gorm:"type:uuid;index"`
	ClassID      uuid.UUID  `json:"class_id" gorm:"type:uuid;not null;index"`
	SectionID    *uuid.UUID `json:"section_id,omitempty" gorm:"type:uuid;index"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateStudentRequest struct {
	FirstName    string `json:"first_name" binding:"required"`
	LastName     string `json:"last_name" binding:"required"`
	ParentUserID string `json:"parent_user_id" binding:"omitempty,uuid"`
	ClassID      string `json:"class_id" binding:"required,uuid"`
	SectionID    string `json:"section_id" binding:"omitempty,uuid"`
}

type UpdateStudentRequest struct {
	FirstName    *string `json:"first_name"`
	LastName     *string `json:"last_name"`
	ParentUserID *string `json:"parent_user_id"`
	ClassID      *string `json:"class_id"`
	SectionID    *string `json:"section_id"`
	IsActive     *bool   `json:"is_active"`
}

type StudentListQuery struct {
	Page         int    `form:"page,default=1"`
	Limit        int    `form:"limit,default=20"`
	Search       string `form:"search"`
	ClassID      string `form:"class_id"`
	SectionID    string `form:"section_id"`
	ParentUserID string `form:"parent_user_id"`
	IsActive     *bool  `form:"is_active"`
}

type StudentListResponse struct {
	Students []Student `json:"students"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}
