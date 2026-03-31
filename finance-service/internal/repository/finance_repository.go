package repository

import (
	"github.com/avaneeshravat/school-management/finance-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FinanceRepository struct {
	db *gorm.DB
}

func NewFinanceRepository(db *gorm.DB) *FinanceRepository {
	return &FinanceRepository{db: db}
}

func (r *FinanceRepository) CreateFee(fee *model.Fee) error {
	return r.db.Create(fee).Error
}

func (r *FinanceRepository) GetFeeByIDAndSchool(feeID, schoolID uuid.UUID) (*model.Fee, error) {
	var fee model.Fee
	err := r.db.Where("id = ? AND school_id = ?", feeID, schoolID).First(&fee).Error
	return &fee, err
}

func (r *FinanceRepository) CreatePayment(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

func (r *FinanceRepository) GetPaidAmountForFeeAndStudent(feeID, studentID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.Model(&model.Payment{}).
		Select("COALESCE(SUM(amount_paid), 0)").
		Where("fee_id = ? AND student_id = ?", feeID, studentID).
		Scan(&total).Error
	return total, err
}

func (r *FinanceRepository) GetFeesForDues(schoolID uuid.UUID, query model.DueQuery) ([]model.Fee, error) {
	var fees []model.Fee
	q := r.db.Where("school_id = ? AND is_active = ?", schoolID, true)

	if query.StudentID != "" {
		q = q.Where("(student_id = ? OR student_id IS NULL)", query.StudentID)
	}
	if query.ClassID != "" {
		q = q.Where("(class_id = ? OR class_id IS NULL)", query.ClassID)
	}
	if query.SectionID != "" {
		q = q.Where("(section_id = ? OR section_id IS NULL)", query.SectionID)
	}

	err := q.Order("created_at desc").Find(&fees).Error
	return fees, err
}
