package model

import (
	"time"

	"github.com/google/uuid"
)

// Vehicle represents a transport vehicle (bus, van, etc.)
type Vehicle struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID        uuid.UUID `json:"school_id" gorm:"type:uuid;not null;index"`
	VehicleNumber   string    `json:"vehicle_number" gorm:"not null"`
	VehicleType     string    `json:"vehicle_type" gorm:"not null"` // bus, van, mini-bus
	Capacity        int       `json:"capacity" gorm:"not null"`
	DriverName      string    `json:"driver_name"`
	DriverPhone     string    `json:"driver_phone"`
	DriverLicense   string    `json:"driver_license"`
	ConductorName   string    `json:"conductor_name"`
	ConductorPhone  string    `json:"conductor_phone"`
	InsuranceExpiry *time.Time `json:"insurance_expiry,omitempty" gorm:"type:date"`
	FitnessExpiry   *time.Time `json:"fitness_expiry,omitempty" gorm:"type:date"`
	IsActive        bool      `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Route represents a transport route
type Route struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID    uuid.UUID `json:"school_id" gorm:"type:uuid;not null;index"`
	RouteName   string    `json:"route_name" gorm:"not null"`
	RouteCode   string    `json:"route_code" gorm:"not null"`
	VehicleID   *uuid.UUID `json:"vehicle_id,omitempty" gorm:"type:uuid;index"`
	StartPoint  string    `json:"start_point"`
	EndPoint    string    `json:"end_point"`
	Distance    float64   `json:"distance"` // in km
	Duration    int       `json:"duration"` // in minutes
	MonthlyFee  float64   `json:"monthly_fee"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Stop represents a pickup/drop stop on a route
type Stop struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID      uuid.UUID `json:"school_id" gorm:"type:uuid;not null;index"`
	RouteID       uuid.UUID `json:"route_id" gorm:"type:uuid;not null;index"`
	StopName      string    `json:"stop_name" gorm:"not null"`
	StopOrder     int       `json:"stop_order" gorm:"not null"`
	PickupTime    string    `json:"pickup_time"`  // e.g., "07:30"
	DropTime      string    `json:"drop_time"`    // e.g., "14:30"
	Landmark      string    `json:"landmark"`
	Latitude      *float64  `json:"latitude,omitempty"`
	Longitude     *float64  `json:"longitude,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// StudentTransport assigns a student to a route/stop
type StudentTransport struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID    uuid.UUID `json:"school_id" gorm:"type:uuid;not null;index"`
	StudentID   uuid.UUID `json:"student_id" gorm:"type:uuid;not null;index"`
	RouteID     uuid.UUID `json:"route_id" gorm:"type:uuid;not null;index"`
	StopID      uuid.UUID `json:"stop_id" gorm:"type:uuid;not null;index"`
	TransportType string  `json:"transport_type" gorm:"not null"` // pickup, drop, both
	StartDate   time.Time `json:"start_date" gorm:"type:date;not null"`
	EndDate     *time.Time `json:"end_date,omitempty" gorm:"type:date"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DTOs for Vehicle
type CreateVehicleRequest struct {
	VehicleNumber   string `json:"vehicle_number" binding:"required"`
	VehicleType     string `json:"vehicle_type" binding:"required"`
	Capacity        int    `json:"capacity" binding:"required,min=1"`
	DriverName      string `json:"driver_name"`
	DriverPhone     string `json:"driver_phone"`
	DriverLicense   string `json:"driver_license"`
	ConductorName   string `json:"conductor_name"`
	ConductorPhone  string `json:"conductor_phone"`
	InsuranceExpiry string `json:"insurance_expiry"` // YYYY-MM-DD
	FitnessExpiry   string `json:"fitness_expiry"`   // YYYY-MM-DD
}

type UpdateVehicleRequest struct {
	VehicleNumber   *string `json:"vehicle_number"`
	VehicleType     *string `json:"vehicle_type"`
	Capacity        *int    `json:"capacity"`
	DriverName      *string `json:"driver_name"`
	DriverPhone     *string `json:"driver_phone"`
	DriverLicense   *string `json:"driver_license"`
	ConductorName   *string `json:"conductor_name"`
	ConductorPhone  *string `json:"conductor_phone"`
	InsuranceExpiry *string `json:"insurance_expiry"`
	FitnessExpiry   *string `json:"fitness_expiry"`
	IsActive        *bool   `json:"is_active"`
}

type VehicleQuery struct {
	Page         int    `form:"page,default=1"`
	Limit        int    `form:"limit,default=20"`
	VehicleType  string `form:"vehicle_type"`
	IsActive     string `form:"is_active"`
}

type VehicleListResponse struct {
	Vehicles []Vehicle `json:"vehicles"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

// DTOs for Route
type CreateRouteRequest struct {
	RouteName   string  `json:"route_name" binding:"required"`
	RouteCode   string  `json:"route_code" binding:"required"`
	VehicleID   string  `json:"vehicle_id"`
	StartPoint  string  `json:"start_point"`
	EndPoint    string  `json:"end_point"`
	Distance    float64 `json:"distance"`
	Duration    int     `json:"duration"`
	MonthlyFee  float64 `json:"monthly_fee"`
}

type UpdateRouteRequest struct {
	RouteName   *string  `json:"route_name"`
	RouteCode   *string  `json:"route_code"`
	VehicleID   *string  `json:"vehicle_id"`
	StartPoint  *string  `json:"start_point"`
	EndPoint    *string  `json:"end_point"`
	Distance    *float64 `json:"distance"`
	Duration    *int     `json:"duration"`
	MonthlyFee  *float64 `json:"monthly_fee"`
	IsActive    *bool    `json:"is_active"`
}

type RouteQuery struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=20"`
	IsActive string `form:"is_active"`
}

type RouteListResponse struct {
	Routes []Route `json:"routes"`
	Total  int64   `json:"total"`
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
}

// DTOs for Stop
type CreateStopRequest struct {
	RouteID    string   `json:"route_id" binding:"required,uuid"`
	StopName   string   `json:"stop_name" binding:"required"`
	StopOrder  int      `json:"stop_order" binding:"required,min=1"`
	PickupTime string   `json:"pickup_time"`
	DropTime   string   `json:"drop_time"`
	Landmark   string   `json:"landmark"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
}

type UpdateStopRequest struct {
	StopName   *string  `json:"stop_name"`
	StopOrder  *int     `json:"stop_order"`
	PickupTime *string  `json:"pickup_time"`
	DropTime   *string  `json:"drop_time"`
	Landmark   *string  `json:"landmark"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
}

type StopQuery struct {
	Page    int    `form:"page,default=1"`
	Limit   int    `form:"limit,default=50"`
	RouteID string `form:"route_id"`
}

type StopListResponse struct {
	Stops []Stop `json:"stops"`
	Total int64  `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

// DTOs for StudentTransport
type CreateStudentTransportRequest struct {
	StudentID     string `json:"student_id" binding:"required,uuid"`
	RouteID       string `json:"route_id" binding:"required,uuid"`
	StopID        string `json:"stop_id" binding:"required,uuid"`
	TransportType string `json:"transport_type" binding:"required"` // pickup, drop, both
	StartDate     string `json:"start_date" binding:"required"`     // YYYY-MM-DD
	EndDate       string `json:"end_date"`
}

type UpdateStudentTransportRequest struct {
	RouteID       *string `json:"route_id"`
	StopID        *string `json:"stop_id"`
	TransportType *string `json:"transport_type"`
	StartDate     *string `json:"start_date"`
	EndDate       *string `json:"end_date"`
	IsActive      *bool   `json:"is_active"`
}

type StudentTransportQuery struct {
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`
	StudentID string `form:"student_id"`
	RouteID   string `form:"route_id"`
	StopID    string `form:"stop_id"`
	IsActive  string `form:"is_active"`
}

type StudentTransportListResponse struct {
	Assignments []StudentTransport `json:"assignments"`
	Total       int64              `json:"total"`
	Page        int                `json:"page"`
	Limit       int                `json:"limit"`
}

// RouteWithDetails includes vehicle and stop count
type RouteWithDetails struct {
	Route
	VehicleNumber string `json:"vehicle_number,omitempty"`
	StopCount     int    `json:"stop_count"`
	StudentCount  int    `json:"student_count"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}
