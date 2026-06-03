package repository

import (
	"github.com/avaneeshravat/school-management/student-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

func (r *StudentRepository) CreateStudent(student *model.Student) error {
	return r.db.Create(student).Error
}

func (r *StudentRepository) GetStudentsBySchoolID(
	schoolID uuid.UUID,
	query model.StudentListQuery,
) ([]model.Student, int64, error) {
	var students []model.Student
	var total int64

	q := r.db.Model(&model.Student{}).Where("school_id = ?", schoolID)

	if query.Search != "" {
		search := "%" + query.Search + "%"
		q = q.Where("first_name ILIKE ? OR last_name ILIKE ?", search, search)
	}
	if query.ClassID != "" {
		q = q.Where("class_id = ?", query.ClassID)
	}
	if query.SectionID != "" {
		q = q.Where("section_id = ?", query.SectionID)
	}
	if query.ParentUserID != "" {
		q = q.Where("parent_user_id = ?", query.ParentUserID)
	}
	if query.IsActive != nil {
		q = q.Where("is_active = ?", *query.IsActive)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.Limit
	err := q.Order("created_at desc").Offset(offset).Limit(query.Limit).Find(&students).Error
	return students, total, err
}

func (r *StudentRepository) GetStudentByIDAndSchoolID(id, schoolID uuid.UUID) (*model.Student, error) {
	var student model.Student
	err := r.db.Where("id = ? AND school_id = ?", id, schoolID).First(&student).Error
	return &student, err
}

func (r *StudentRepository) GetStudentByID(id uuid.UUID) (*model.Student, error) {
	var student model.Student
	err := r.db.Where("id = ?", id).First(&student).Error
	return &student, err
}

func (r *StudentRepository) UpdateStudent(student *model.Student) error {
	return r.db.Save(student).Error
}

// DeleteStudent removes a student row by id. Used to roll back admission if
// downstream provisioning (e.g. login creation) fails.
func (r *StudentRepository) DeleteStudent(id uuid.UUID) error {
	return r.db.Delete(&model.Student{}, "id = ?", id).Error
}
