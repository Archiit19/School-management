package repository

import (
	"github.com/avaneeshravat/school-management/auth-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// ─── School ─────────────────────────────────────────────────────────

func (r *AuthRepository) CreateSchool(school *model.School) error {
	return r.db.Create(school).Error
}

func (r *AuthRepository) GetSchoolByEmail(email string) (*model.School, error) {
	var school model.School
	err := r.db.Where("email = ?", email).First(&school).Error
	return &school, err
}

// ─── User ───────────────────────────────────────────────────────────

func (r *AuthRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *AuthRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *AuthRepository) GetUserByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	return &user, err
}

// ─── User Management (Flow 2) ──────────────────────────────────────

func (r *AuthRepository) GetUsersBySchoolID(schoolID uuid.UUID, query model.UserListQuery) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	db := r.db.Where("school_id = ?", schoolID)

	// Filter by role
	if query.RoleID != "" {
		roleID, err := uuid.Parse(query.RoleID)
		if err == nil {
			db = db.Where("role_id = ?", roleID)
		}
	}

	// Filter by active status
	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}

	// Search by name or email
	if query.Search != "" {
		search := "%" + query.Search + "%"
		db = db.Where("name ILIKE ? OR email ILIKE ?", search, search)
	}

	// Count total before pagination
	if err := db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (query.Page - 1) * query.Limit
	if err := db.Offset(offset).Limit(query.Limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *AuthRepository) GetUserByIDAndSchool(id uuid.UUID, schoolID uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ? AND school_id = ?", id, schoolID).First(&user).Error
	return &user, err
}

func (r *AuthRepository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *AuthRepository) DeleteUser(id uuid.UUID, schoolID uuid.UUID) error {
	return r.db.Where("id = ? AND school_id = ?", id, schoolID).Delete(&model.User{}).Error
}
