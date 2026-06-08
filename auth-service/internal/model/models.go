package model

import (
	"time"

	"github.com/google/uuid"
)

// School is the school record returned by school-service (not stored in auth DB).
type School struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string    `json:"name" example:"Springfield Elementary"`
	Address   string    `json:"address" example:"123 Main St"`
	Phone     string    `json:"phone" example:"555-0100"`
	Email     string    `json:"email" example:"info@springfield.edu"`
	IsActive  bool      `json:"is_active" example:"true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User represents an authenticated user in the system.
// School and role context live in school-service user_schools mapping (JWT / API only here).
type User struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()" example:"550e8400-e29b-41d4-a716-446655440001"`
	Name        string     `json:"name" gorm:"not null" example:"John Doe"`
	Email       string     `json:"email" gorm:"uniqueIndex;not null" example:"john@springfield.edu"`
	Password    string     `json:"-" gorm:"not null"`
	StudentID   *uuid.UUID `json:"student_id,omitempty" gorm:"type:uuid;index" example:"550e8400-e29b-41d4-a716-446655440099"`
	SchoolID    *uuid.UUID `json:"school_id,omitempty" gorm:"-" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoleID      *uuid.UUID `json:"role_id,omitempty" gorm:"-" example:"550e8400-e29b-41d4-a716-446655440002"`
	RoleName    string     `json:"role_name" gorm:"-" example:"platform_admin"`
	Permissions []string   `json:"permissions,omitempty" gorm:"-"`
	School      *School    `json:"school,omitempty" gorm:"-"`
	Schools     []School   `json:"schools,omitempty" gorm:"-"`
	IsActive    bool       `json:"is_active" gorm:"default:true" example:"true"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ─── Request / Response DTOs ────────────────────────────────────────

// SignupRequest creates a platform admin account (no school yet).
type SignupRequest struct {
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"secret123"`
}

// UpdateProfileRequest updates the authenticated user's profile.
type UpdateProfileRequest struct {
	Name  *string `json:"name" example:"Jane Doe"`
	Email *string `json:"email" example:"jane@example.com"`
}

// SelectSchoolRequest enters a school admin context for a mapped school.
type SelectSchoolRequest struct {
	SchoolID string `json:"school_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// RegisterSchoolRequest is the legacy payload for registering a new school with its super admin.
type RegisterSchoolRequest struct {
	SchoolName    string `json:"school_name" binding:"required" example:"Springfield Elementary"`
	SchoolAddress string `json:"school_address" example:"123 Main St"`
	SchoolPhone   string `json:"school_phone" example:"555-0100"`
	SchoolEmail   string `json:"school_email" binding:"required,email" example:"info@springfield.edu"`
	AdminName     string `json:"admin_name" binding:"required" example:"John Doe"`
	AdminEmail    string `json:"admin_email" binding:"required,email" example:"john@springfield.edu"`
	AdminPassword string `json:"admin_password" binding:"required,min=6" example:"secret123"`
}

// LoginRequest is the payload for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@springfield.edu"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

// LoginResponse contains the JWT token and user info returned on login.
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	User  User   `json:"user"`
}

// RegisterSchoolResponse contains the school, admin user, and JWT token.
type RegisterSchoolResponse struct {
	School School `json:"school"`
	Admin  User   `json:"admin"`
	Token  string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

// ─── User Management DTOs (Flow 2) ─────────────────────────────────

// CreateUserRequest is the payload for creating a new user (admin only).
type CreateUserRequest struct {
	Name      string `json:"name" binding:"required" example:"Jane Smith"`
	Email     string `json:"email" binding:"required,email" example:"jane@springfield.edu"`
	Password  string `json:"password" binding:"required,min=6" example:"teacher123"`
	RoleID    string `json:"role_id" binding:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440002"`
	StudentID string `json:"student_id,omitempty" binding:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440099"`
}

// UpdateUserRequest is the payload for partially updating a user.
type UpdateUserRequest struct {
	Name      *string `json:"name" example:"Jane Doe"`
	Email     *string `json:"email" example:"jane.doe@springfield.edu"`
	RoleID    *string `json:"role_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	StudentID *string `json:"student_id" example:"550e8400-e29b-41d4-a716-446655440099"`
	IsActive  *bool   `json:"is_active" example:"false"`
}

// UserListQuery holds query parameters for listing users.
type UserListQuery struct {
	Page     int    `form:"page,default=1" example:"1"`
	Limit    int    `form:"limit,default=20" example:"20"`
	Search   string `form:"search" example:"jane"`
	RoleID   string `form:"role_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	IsActive *bool  `form:"is_active" example:"true"`
}

// UserListResponse is the paginated list of users.
type UserListResponse struct {
	Users []User `json:"users"`
	Total int64  `json:"total" example:"3"`
	Page  int    `json:"page" example:"1"`
	Limit int    `json:"limit" example:"20"`
}

// ErrorResponse represents a generic error response.
type ErrorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}

// MessageResponse represents a generic success message.
type MessageResponse struct {
	Message string `json:"message" example:"operation successful"`
}

// CreateStudentLoginRequest is used by student-service when admitting a pupil with a login.
type CreateStudentLoginRequest struct {
	SchoolID  string `json:"school_id" binding:"required,uuid"`
	StudentID string `json:"student_id" binding:"required,uuid"`
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
}
