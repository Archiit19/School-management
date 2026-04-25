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

type BulkAttendanceEntry struct {
	StudentID string `json:"student_id" binding:"required,uuid"`
	Status    string `json:"status" binding:"required"`
	Remarks   string `json:"remarks"`
}

type BulkCreateAttendanceRequest struct {
	ClassID   string                `json:"class_id" binding:"required,uuid"`
	SectionID string                `json:"section_id" binding:"omitempty,uuid"`
	SubjectID string                `json:"subject_id" binding:"omitempty,uuid"`
	Date      string                `json:"date" binding:"required"`
	Entries   []BulkAttendanceEntry `json:"entries" binding:"required,min=1,dive"`
}

type BulkAttendanceResponse struct {
	Created int          `json:"created"`
	Skipped int          `json:"skipped"`
	Records []Attendance `json:"records"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}

// TeacherAttendance records daily presence for staff (teachers); separate from student Attendance.
type TeacherAttendance struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID         uuid.UUID `json:"school_id" gorm:"type:uuid;not null;index"`
	TeacherUserID    uuid.UUID `json:"teacher_user_id" gorm:"type:uuid;not null;index"`
	RecordedByUserID uuid.UUID `json:"recorded_by_user_id" gorm:"type:uuid;not null;index"`
	Date             time.Time `json:"date" gorm:"type:date;not null;index"`
	Status           string    `json:"status" gorm:"not null"`
	Remarks          string    `json:"remarks"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateTeacherAttendanceRequest struct {
	TeacherUserID string `json:"teacher_user_id" binding:"omitempty,uuid"` // omit to mark own attendance
	Date          string `json:"date" binding:"required"`
	Status        string `json:"status" binding:"required"`
	Remarks       string `json:"remarks"`
}

type UpdateTeacherAttendanceRequest struct {
	Status  *string `json:"status"`
	Remarks *string `json:"remarks"`
}

type TeacherAttendanceQuery struct {
	Page          int    `form:"page,default=1"`
	Limit         int    `form:"limit,default=20"`
	Date          string `form:"date"`
	TeacherUserID string `form:"teacher_user_id"`
	Status        string `form:"status"`
}

type TeacherAttendanceListResponse struct {
	Attendance []TeacherAttendance `json:"attendance"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
}

type BulkTeacherAttendanceEntry struct {
	TeacherUserID string `json:"teacher_user_id" binding:"required,uuid"`
	Status        string `json:"status" binding:"required"`
	Remarks       string `json:"remarks"`
}

type BulkCreateTeacherAttendanceRequest struct {
	Date    string                       `json:"date" binding:"required"`
	Entries []BulkTeacherAttendanceEntry `json:"entries" binding:"required,min=1,dive"`
}

type BulkTeacherAttendanceResponse struct {
	Created int                 `json:"created"`
	Skipped int                 `json:"skipped"`
	Records []TeacherAttendance `json:"records"`
}

// AttendanceStatsQuery filters for calculating attendance statistics.
type AttendanceStatsQuery struct {
	StudentID string `form:"student_id" binding:"omitempty,uuid"`
	ClassID   string `form:"class_id" binding:"omitempty,uuid"`
	SectionID string `form:"section_id" binding:"omitempty,uuid"`
	SubjectID string `form:"subject_id" binding:"omitempty,uuid"`
	StartDate string `form:"start_date"` // YYYY-MM-DD
	EndDate   string `form:"end_date"`   // YYYY-MM-DD
}

// AttendanceStats contains calculated attendance percentages.
type AttendanceStats struct {
	TotalDays       int     `json:"total_days"`
	PresentDays     int     `json:"present_days"`
	AbsentDays      int     `json:"absent_days"`
	LateDays        int     `json:"late_days"`
	ExcusedDays     int     `json:"excused_days"`
	AttendanceRate  float64 `json:"attendance_rate"`  // (present+late+excused)/total * 100
	StudentID       string  `json:"student_id,omitempty"`
	ClassID         string  `json:"class_id,omitempty"`
}

// AttendanceStatsResponse wraps stats with context info.
type AttendanceStatsResponse struct {
	Stats     []AttendanceStats `json:"stats"`
	StartDate string            `json:"start_date"`
	EndDate   string            `json:"end_date"`
}

// TeacherAttendanceStatsQuery filters for teacher stats.
type TeacherAttendanceStatsQuery struct {
	TeacherUserID string `form:"teacher_user_id" binding:"omitempty,uuid"`
	StartDate     string `form:"start_date"` // YYYY-MM-DD
	EndDate       string `form:"end_date"`   // YYYY-MM-DD
}

// TeacherAttendanceStats contains calculated teacher attendance percentages.
type TeacherAttendanceStats struct {
	TotalDays       int     `json:"total_days"`
	PresentDays     int     `json:"present_days"`
	AbsentDays      int     `json:"absent_days"`
	LateDays        int     `json:"late_days"`
	ExcusedDays     int     `json:"excused_days"`
	AttendanceRate  float64 `json:"attendance_rate"`
	TeacherUserID   string  `json:"teacher_user_id,omitempty"`
}

// TeacherAttendanceStatsResponse wraps teacher stats.
type TeacherAttendanceStatsResponse struct {
	Stats     []TeacherAttendanceStats `json:"stats"`
	StartDate string                   `json:"start_date"`
	EndDate   string                   `json:"end_date"`
}
