package model

import (
	"time"

	"github.com/google/uuid"
)

type Class struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID    uuid.UUID `json:"school_id" gorm:"type:uuid;not null;index"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Section struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID  uuid.UUID `json:"school_id" gorm:"type:uuid;not null;index"`
	ClassID   uuid.UUID `json:"class_id" gorm:"type:uuid;not null;index"`
	Name      string    `json:"name" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Subject struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID  uuid.UUID  `json:"school_id" gorm:"type:uuid;not null;index"`
	ClassID   uuid.UUID  `json:"class_id" gorm:"type:uuid;not null;index"`
	SectionID *uuid.UUID `json:"section_id,omitempty" gorm:"type:uuid;index"`
	Name      string     `json:"name" gorm:"not null"`
	Code      string     `json:"code"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type TeacherAssignment struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID      uuid.UUID `json:"school_id" gorm:"type:uuid;not null;index"`
	TeacherUserID uuid.UUID `json:"teacher_user_id" gorm:"type:uuid;not null;index"`
	ClassID       uuid.UUID `json:"class_id" gorm:"type:uuid;not null;index"`
	SubjectID     uuid.UUID `json:"subject_id" gorm:"type:uuid;not null;index"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Assignment struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID      uuid.UUID  `json:"school_id" gorm:"type:uuid;not null;index"`
	TeacherUserID uuid.UUID  `json:"teacher_user_id" gorm:"type:uuid;not null;index"`
	ClassID       uuid.UUID  `json:"class_id" gorm:"type:uuid;not null;index"`
	SubjectID     uuid.UUID  `json:"subject_id" gorm:"type:uuid;not null;index"`
	Title         string     `json:"title" gorm:"not null"`
	Description   string     `json:"description"`
	MaterialURL   string     `json:"material_url"`
	DueDate       *time.Time `json:"due_date,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Submission struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID        uuid.UUID  `json:"school_id" gorm:"type:uuid;not null;index"`
	AssignmentID    uuid.UUID  `json:"assignment_id" gorm:"type:uuid;not null;index"`
	StudentID       uuid.UUID  `json:"student_id" gorm:"type:uuid;not null;index"`
	SubmittedBy     uuid.UUID  `json:"submitted_by" gorm:"type:uuid;not null;index"`
	Content         string     `json:"content"`
	MaterialURL     string     `json:"material_url"`
	TeacherFeedback string     `json:"teacher_feedback"`
	Marks           *int       `json:"marks,omitempty" gorm:"type:integer"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// SubmissionView is returned when teachers list submissions for an assignment.
type SubmissionView struct {
	Submission
	StudentName   string `json:"student_name,omitempty"`
	StudentCode   string `json:"student_code,omitempty"`
	SubmitterName string `json:"submitter_name,omitempty"`
}

type CreateClassRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type CreateSectionRequest struct {
	ClassID string `json:"class_id" binding:"required,uuid"`
	Name    string `json:"name" binding:"required"`
}

type CreateSubjectRequest struct {
	ClassID   string `json:"class_id" binding:"required,uuid"`
	SectionID string `json:"section_id,omitempty" binding:"omitempty,uuid"`
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code"`
}

type ClassWithChildren struct {
	Class    Class     `json:"class"`
	Sections []Section `json:"sections"`
	Subjects []Subject `json:"subjects"`
}

type CreateTeacherAssignmentRequest struct {
	TeacherUserID string `json:"teacher_user_id" binding:"required,uuid"`
	ClassID       string `json:"class_id" binding:"required,uuid"`
	SubjectID     string `json:"subject_id" binding:"required,uuid"`
}

type UpdateTeacherAssignmentRequest struct {
	TeacherUserID *string `json:"teacher_user_id" binding:"omitempty,uuid"`
	ClassID       *string `json:"class_id" binding:"omitempty,uuid"`
	SubjectID     *string `json:"subject_id" binding:"omitempty,uuid"`
}

type TeacherAssignmentQuery struct {
	TeacherUserID string `form:"teacher_user_id"`
	ClassID       string `form:"class_id"`
	SubjectID     string `form:"subject_id"`
}

type CreateAssignmentRequest struct {
	ClassID     string `json:"class_id" binding:"required,uuid"`
	SubjectID   string `json:"subject_id" binding:"required,uuid"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	MaterialURL string `json:"material_url"`
	DueDate     string `json:"due_date"`
}

type AssignmentQuery struct {
	ClassID   string `form:"class_id"`
	SubjectID string `form:"subject_id"`
	TeacherID string `form:"teacher_id"`
}

type CreateSubmissionRequest struct {
	AssignmentID string `json:"assignment_id" binding:"required,uuid"`
	StudentID    string `json:"student_id" binding:"required,uuid"`
	Content      string `json:"content"`
	MaterialURL  string `json:"material_url"`
}

// CreateMySubmissionRequest is the pupil-portal payload — student_id is forced from JWT.
type CreateMySubmissionRequest struct {
	AssignmentID string `json:"assignment_id" binding:"required,uuid"`
	Content      string `json:"content"`
	MaterialURL  string `json:"material_url"`
}

type UpdateSubmissionRequest struct {
	TeacherFeedback *string `json:"teacher_feedback"`
	Marks           *int    `json:"marks"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}

// MyAcademicProfile is the consolidated academic context for the authenticated pupil:
// their class + section, the subjects taught in that class, and the teachers assigned.
type MyAcademicProfile struct {
	Class    *Class           `json:"class,omitempty"`
	Section  *Section         `json:"section,omitempty"`
	Subjects []Subject        `json:"subjects"`
	Teachers []ClassTeacher   `json:"teachers"`
}

// ClassTeacher is one teacher assigned to one subject in the pupil's class.
type ClassTeacher struct {
	TeacherUserID uuid.UUID `json:"teacher_user_id"`
	TeacherName   string    `json:"teacher_name"`
	TeacherEmail  string    `json:"teacher_email,omitempty"`
	SubjectID     uuid.UUID `json:"subject_id"`
	SubjectName   string    `json:"subject_name"`
}

// StudentEnrollment maps a pupil (user_id) to class/section for efficient lookups.
type StudentEnrollment struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID  uuid.UUID  `json:"school_id" gorm:"type:uuid;not null;uniqueIndex:idx_enrollment_school_user"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_enrollment_school_user;index"`
	ClassID   uuid.UUID  `json:"class_id" gorm:"type:uuid;not null;index:idx_enrollment_class_section,priority:1"`
	SectionID *uuid.UUID `json:"section_id,omitempty" gorm:"type:uuid;index:idx_enrollment_class_section,priority:2"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type EnrollmentQuery struct {
	ClassID   string `form:"class_id" binding:"required,uuid"`
	SectionID string `form:"section_id" binding:"omitempty,uuid"`
}

type EnrollmentListResponse struct {
	Enrollments []StudentEnrollment `json:"enrollments"`
	Total       int64               `json:"total"`
}

type UpsertEnrollmentRequest struct {
	UserID    string `json:"user_id" binding:"required,uuid"`
	SchoolID  string `json:"school_id" binding:"required,uuid"`
	ClassID   string `json:"class_id" binding:"required,uuid"`
	SectionID string `json:"section_id" binding:"omitempty,uuid"`
}

type UpdateEnrollmentRequest struct {
	ClassID   string `json:"class_id" binding:"required,uuid"`
	SectionID string `json:"section_id" binding:"omitempty,uuid"`
}
