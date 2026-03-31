package model

import (
	"time"

	"github.com/google/uuid"
)

type Exam struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID    uuid.UUID  `json:"school_id" gorm:"type:uuid;not null;index"`
	ClassID     uuid.UUID  `json:"class_id" gorm:"type:uuid;not null;index"`
	SectionID   *uuid.UUID `json:"section_id,omitempty" gorm:"type:uuid;index"`
	SubjectID   *uuid.UUID `json:"subject_id,omitempty" gorm:"type:uuid;index"`
	Title       string     `json:"title" gorm:"not null"`
	ExamDate    time.Time  `json:"exam_date" gorm:"type:date;not null"`
	TotalMarks  float64    `json:"total_marks" gorm:"not null"`
	IsPublished bool       `json:"is_published" gorm:"default:false"`
	CreatedBy   uuid.UUID  `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Mark struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID      uuid.UUID `json:"school_id" gorm:"type:uuid;not null;index"`
	ExamID        uuid.UUID `json:"exam_id" gorm:"type:uuid;not null;index"`
	StudentID     uuid.UUID `json:"student_id" gorm:"type:uuid;not null;index"`
	MarksObtained float64   `json:"marks_obtained" gorm:"not null"`
	Remarks       string    `json:"remarks"`
	CreatedBy     uuid.UUID `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateExamRequest struct {
	ClassID    string  `json:"class_id" binding:"required,uuid"`
	SectionID  string  `json:"section_id" binding:"omitempty,uuid"`
	SubjectID  string  `json:"subject_id" binding:"omitempty,uuid"`
	Title      string  `json:"title" binding:"required"`
	ExamDate   string  `json:"exam_date" binding:"required"`
	TotalMarks float64 `json:"total_marks" binding:"required,gt=0"`
}

type EnterMarksRequest struct {
	ExamID        string  `json:"exam_id" binding:"required,uuid"`
	StudentID     string  `json:"student_id" binding:"required,uuid"`
	MarksObtained float64 `json:"marks_obtained" binding:"required"`
	Remarks       string  `json:"remarks"`
}

type PublishResultRequest struct {
	ExamID string `json:"exam_id" binding:"required,uuid"`
}

type ResultQuery struct {
	ExamID    string `form:"exam_id"`
	StudentID string `form:"student_id"`
	ClassID   string `form:"class_id"`
}

type ResultItem struct {
	ExamID        uuid.UUID `json:"exam_id"`
	ExamTitle     string    `json:"exam_title"`
	ExamDate      time.Time `json:"exam_date"`
	StudentID     uuid.UUID `json:"student_id"`
	MarksObtained float64   `json:"marks_obtained"`
	TotalMarks    float64   `json:"total_marks"`
	Percentage    float64   `json:"percentage"`
	Grade         string    `json:"grade"`
	Published     bool      `json:"published"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}
