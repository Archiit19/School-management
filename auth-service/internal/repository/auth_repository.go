package repository

import (
	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

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

func (r *AuthRepository) GetUsersByIDs(userIDs []uuid.UUID, query model.UserListQuery) ([]model.User, int64, error) {
	if len(userIDs) == 0 {
		return []model.User{}, 0, nil
	}

	var users []model.User
	var total int64

	db := r.db.Where("id IN ?", userIDs)

	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}

	if query.Search != "" {
		search := "%" + query.Search + "%"
		db = db.Where("name ILIKE ? OR email ILIKE ?", search, search)
	}

	if err := db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.Limit
	if err := db.Offset(offset).Limit(query.Limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *AuthRepository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *AuthRepository) DeleteUser(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.User{}).Error
}
