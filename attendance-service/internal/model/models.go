package model

import (
	"time"

	"github.com/google/uuid"
)

type Attendance struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID      uuid.UUID  `json:"school_id" gorm:"type:uuid;not null;index"`
	TeacherUserID uuid.UUID  `json:"teacher_user_id" gorm:"type:uuid;not null;index"`
	StudentID     uuid.UUID  `json:"student_id" gorm:"type:uuid;not null;index"`
	ClassID       uuid.UUID  `json:"class_id" gorm:"type:uuid;not null;index"`
	SectionID     *uuid.UUID `json:"section_id,omitempty" gorm:"type:uuid;index"`
	SubjectID     *uuid.UUID `json:"subject_id,omitempty" gorm:"type:uuid;index"`
	Date          time.Time  `json:"date" gorm:"type:date;not null;index"`
	Status        string     `json:"status" gorm:"not null"`
	Remarks       string     `json:"remarks"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CreateAttendanceRequest struct {
	StudentID string `json:"student_id" binding:"required,uuid"`
	ClassID   string `json:"class_id" binding:"required,uuid"`
	SectionID string `json:"section_id" binding:"omitempty,uuid"`
	SubjectID string `json:"subject_id" binding:"omitempty,uuid"`
	Date      string `json:"date" binding:"required"`
	Status    string `json:"status" binding:"required"`
	Remarks   string `json:"remarks"`
}

type UpdateAttendanceRequest struct {
	Status  *string `json:"status"`
	Remarks *string `json:"remarks"`
}

type AttendanceQuery struct {
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`
	Date      string `form:"date"`
	StudentID string `form:"student_id"`
	ClassID   string `form:"class_id"`
	SectionID string `form:"section_id"`
	SubjectID string `form:"subject_id"`
	Status    string `form:"status"`
}

type AttendanceListResponse struct {
	Attendance []Attendance `json:"attendance"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}
