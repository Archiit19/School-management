package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/avaneeshravat/school-management/finance-service/internal/model"
	"github.com/avaneeshravat/school-management/finance-service/internal/repository"
	"github.com/google/uuid"
)

type FinanceService struct {
	repo *repository.FinanceRepository
}

func NewFinanceService(repo *repository.FinanceRepository) *FinanceService {
	return &FinanceService{repo: repo}
}

func (s *FinanceService) CreateFee(
	req model.CreateFeeRequest,
	schoolID, createdBy uuid.UUID,
	roleName string,
) (*model.Fee, error) {
	classID, err := parseOptionalUUID(req.ClassID, "class_id")
	if err != nil {
		return nil, err
	}
	sectionID, err := parseOptionalUUID(req.SectionID, "section_id")
	if err != nil {
		return nil, err
	}
	studentID, err := parseOptionalUUID(req.StudentID, "student_id")
	if err != nil {
		return nil, err
	}

	var dueDate *time.Time
	if strings.TrimSpace(req.DueDate) != "" {
		parsedDate, err := time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			return nil, errors.New("invalid due_date format, use YYYY-MM-DD")
		}
		dueDate = &parsedDate
	}

	fee := &model.Fee{
		SchoolID:    schoolID,
		Title:       req.Title,
		Description: req.Description,
		Amount:      req.Amount,
		ClassID:     classID,
		SectionID:   sectionID,
		StudentID:   studentID,
		DueDate:     dueDate,
		IsActive:    true,
		CreatedBy:   createdBy,
	}
	if err := s.repo.CreateFee(fee); err != nil {
		return nil, fmt.Errorf("failed to create fee: %w", err)
	}
	return fee, nil
}

func (s *FinanceService) RecordPayment(
	req model.CreatePaymentRequest,
	schoolID, receivedBy uuid.UUID,
	roleName string,
) (*model.Payment, error) {
	feeID, err := uuid.Parse(req.FeeID)
	if err != nil {
		return nil, errors.New("invalid fee_id")
	}
	studentID, err := uuid.Parse(req.StudentID)
	if err != nil {
		return nil, errors.New("invalid student_id")
	}

	fee, err := s.repo.GetFeeByIDAndSchool(feeID, schoolID)
	if err != nil {
		return nil, errors.New("fee not found")
	}

	// If fee is assigned to a specific student, payment must match.
	if fee.StudentID != nil && *fee.StudentID != studentID {
		return nil, errors.New("payment student does not match assigned student for this fee")
	}

	paidSoFar, err := s.repo.GetPaidAmountForFeeAndStudent(feeID, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate paid amount: %w", err)
	}
	if paidSoFar+req.AmountPaid > fee.Amount {
		return nil, errors.New("payment exceeds fee amount")
	}

	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, errors.New("invalid payment_date format, use YYYY-MM-DD")
	}

	payment := &model.Payment{
		SchoolID:    schoolID,
		FeeID:       feeID,
		StudentID:   studentID,
		AmountPaid:  req.AmountPaid,
		PaymentDate: paymentDate,
		Method:      req.Method,
		Reference:   req.Reference,
		ReceivedBy:  receivedBy,
	}
	if err := s.repo.CreatePayment(payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}
	return payment, nil
}

func (s *FinanceService) GetDues(schoolID uuid.UUID, query model.DueQuery) ([]model.DueItem, error) {
	fees, err := s.repo.GetFeesForDues(schoolID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fees: %w", err)
	}

	dues := make([]model.DueItem, 0, len(fees))
	for _, fee := range fees {
		var targetStudentID uuid.UUID
		if fee.StudentID != nil {
			targetStudentID = *fee.StudentID
		} else if query.StudentID != "" {
			parsed, err := uuid.Parse(query.StudentID)
			if err != nil {
				return nil, errors.New("invalid student_id in query")
			}
			targetStudentID = parsed
		} else {
			// For class-level dues, when no student_id is provided, report raw fee amount as pending.
			dues = append(dues, model.DueItem{
				FeeID:      fee.ID,
				Title:      fee.Title,
				Amount:     fee.Amount,
				PaidAmount: 0,
				Balance:    fee.Amount,
				Status:     "pending",
				DueDate:    fee.DueDate,
				StudentID:  fee.StudentID,
				ClassID:    fee.ClassID,
				SectionID:  fee.SectionID,
			})
			continue
		}

		paid, err := s.repo.GetPaidAmountForFeeAndStudent(fee.ID, targetStudentID)
		if err != nil {
			return nil, fmt.Errorf("failed to compute due for fee %s: %w", fee.ID, err)
		}

		balance := fee.Amount - paid
		if balance < 0 {
			balance = 0
		}

		status := "pending"
		if balance == 0 {
			status = "paid"
		} else if paid > 0 {
			status = "partial"
		}

		studentID := fee.StudentID
		if studentID == nil {
			id := targetStudentID
			studentID = &id
		}

		dues = append(dues, model.DueItem{
			FeeID:      fee.ID,
			Title:      fee.Title,
			Amount:     fee.Amount,
			PaidAmount: paid,
			Balance:    balance,
			Status:     status,
			DueDate:    fee.DueDate,
			StudentID:  studentID,
			ClassID:    fee.ClassID,
			SectionID:  fee.SectionID,
		})
	}

	return dues, nil
}

func parseOptionalUUID(value, fieldName string) (*uuid.UUID, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("invalid %s", fieldName)
	}
	return &parsed, nil
}
