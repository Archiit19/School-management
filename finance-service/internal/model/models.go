package model

import (
	"time"

	"github.com/google/uuid"
)

type Fee struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID    uuid.UUID  `json:"school_id" gorm:"type:uuid;not null;index"`
	Title       string     `json:"title" gorm:"not null"`
	Description string     `json:"description"`
	Amount      float64    `json:"amount" gorm:"not null"`
	ClassID     *uuid.UUID `json:"class_id,omitempty" gorm:"type:uuid;index"`
	SectionID   *uuid.UUID `json:"section_id,omitempty" gorm:"type:uuid;index"`
	StudentID   *uuid.UUID `json:"student_id,omitempty" gorm:"type:uuid;index"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedBy   uuid.UUID  `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Payment struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SchoolID    uuid.UUID `json:"school_id" gorm:"type:uuid;not null;index"`
	FeeID       uuid.UUID `json:"fee_id" gorm:"type:uuid;not null;index"`
	StudentID   uuid.UUID `json:"student_id" gorm:"type:uuid;not null;index"`
	AmountPaid  float64   `json:"amount_paid" gorm:"not null"`
	PaymentDate time.Time `json:"payment_date" gorm:"type:date;not null"`
	Method      string    `json:"method"`
	Reference   string    `json:"reference"`
	ReceivedBy  uuid.UUID `json:"received_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateFeeRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	ClassID     string  `json:"class_id" binding:"omitempty,uuid"`
	SectionID   string  `json:"section_id" binding:"omitempty,uuid"`
	StudentID   string  `json:"student_id" binding:"omitempty,uuid"`
	DueDate     string  `json:"due_date"`
}

type CreatePaymentRequest struct {
	FeeID       string  `json:"fee_id" binding:"required,uuid"`
	StudentID   string  `json:"student_id" binding:"required,uuid"`
	AmountPaid  float64 `json:"amount_paid" binding:"required,gt=0"`
	PaymentDate string  `json:"payment_date" binding:"required"`
	Method      string  `json:"method"`
	Reference   string  `json:"reference"`
}

type DueQuery struct {
	StudentID string `form:"student_id"`
	ClassID   string `form:"class_id"`
	SectionID string `form:"section_id"`
}

type DueItem struct {
	FeeID      uuid.UUID  `json:"fee_id"`
	Title      string     `json:"title"`
	Amount     float64    `json:"amount"`
	PaidAmount float64    `json:"paid_amount"`
	Balance    float64    `json:"balance"`
	Status     string     `json:"status"`
	DueDate    *time.Time `json:"due_date,omitempty"`
	StudentID  *uuid.UUID `json:"student_id,omitempty"`
	ClassID    *uuid.UUID `json:"class_id,omitempty"`
	SectionID  *uuid.UUID `json:"section_id,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}
