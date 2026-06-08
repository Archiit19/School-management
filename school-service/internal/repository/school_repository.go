package repository

import (
	"github.com/Archiit19/School-management/school-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SchoolRepository struct {
	db *gorm.DB
}

func NewSchoolRepository(db *gorm.DB) *SchoolRepository {
	return &SchoolRepository{db: db}
}

func (r *SchoolRepository) Create(school *model.School) error {
	return r.db.Create(school).Error
}

func (r *SchoolRepository) GetByID(id uuid.UUID) (*model.School, error) {
	var school model.School
	err := r.db.Where("id = ?", id).First(&school).Error
	return &school, err
}

func (r *SchoolRepository) GetByEmail(email string) (*model.School, error) {
	var school model.School
	err := r.db.Where("email = ?", email).First(&school).Error
	return &school, err
}

func (r *SchoolRepository) Update(school *model.School) error {
	return r.db.Save(school).Error
}

func (r *SchoolRepository) List(query model.SchoolListQuery) ([]model.School, int64, error) {
	var schools []model.School
	var total int64

	db := r.db.Model(&model.School{})
	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}
	if query.Search != "" {
		search := "%" + query.Search + "%"
		db = db.Where("name ILIKE ? OR email ILIKE ?", search, search)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.Limit
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Page <= 0 {
		query.Page = 1
	}

	err := db.Offset(offset).Limit(query.Limit).Order("created_at DESC").Find(&schools).Error
	return schools, total, err
}

func (r *SchoolRepository) CreateMembership(m *model.UserSchool) error {
	return r.db.Create(m).Error
}

func (r *SchoolRepository) GetMembership(schoolID, userID uuid.UUID) (*model.UserSchool, error) {
	var m model.UserSchool
	err := r.db.Where("school_id = ? AND user_id = ?", schoolID, userID).First(&m).Error
	return &m, err
}

func (r *SchoolRepository) GetMembershipsForUser(userID uuid.UUID) ([]model.UserSchool, error) {
	var rows []model.UserSchool
	err := r.db.Where("user_id = ?", userID).Order("created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *SchoolRepository) GetMembersForSchool(schoolID uuid.UUID) ([]model.UserSchool, error) {
	var rows []model.UserSchool
	err := r.db.Where("school_id = ?", schoolID).Order("created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *SchoolRepository) UpdateMembershipRole(schoolID, userID, roleID uuid.UUID) error {
	return r.db.Model(&model.UserSchool{}).
		Where("school_id = ? AND user_id = ?", schoolID, userID).
		Update("role_id", roleID).Error
}

func (r *SchoolRepository) DeleteMembership(schoolID, userID uuid.UUID) error {
	return r.db.Where("school_id = ? AND user_id = ?", schoolID, userID).Delete(&model.UserSchool{}).Error
}

func (r *SchoolRepository) IsUserMember(schoolID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.UserSchool{}).
		Where("school_id = ? AND user_id = ?", schoolID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *SchoolRepository) ListSchoolsForUser(userID uuid.UUID) ([]model.School, error) {
	var schools []model.School
	err := r.db.
		Joins("JOIN user_schools ON user_schools.school_id = schools.id").
		Where("user_schools.user_id = ?", userID).
		Order("schools.created_at DESC").
		Find(&schools).Error
	return schools, err
}

func (r *SchoolRepository) ListUserIDsForSchool(schoolID uuid.UUID, roleID *uuid.UUID) ([]uuid.UUID, error) {
	q := r.db.Model(&model.UserSchool{}).Where("school_id = ?", schoolID)
	if roleID != nil {
		q = q.Where("role_id = ?", *roleID)
	}
	var ids []uuid.UUID
	err := q.Pluck("user_id", &ids).Error
	return ids, err
}
