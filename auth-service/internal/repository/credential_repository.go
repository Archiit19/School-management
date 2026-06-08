package repository

import (
	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CredentialRepository struct {
	db *gorm.DB
}

func NewCredentialRepository(db *gorm.DB) *CredentialRepository {
	return &CredentialRepository{db: db}
}

func (r *CredentialRepository) SetCredential(cred *model.UserCredential) error {
	return r.db.Save(cred).Error
}

func (r *CredentialRepository) GetCredentialByUserID(userID uuid.UUID) (*model.UserCredential, error) {
	var cred model.UserCredential
	err := r.db.Where("user_id = ?", userID).First(&cred).Error
	return &cred, err
}

func (r *CredentialRepository) DeleteCredential(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.UserCredential{}).Error
}

func (r *CredentialRepository) AssignUserRole(ur *model.UserRole) error {
	return r.db.Create(ur).Error
}

func (r *CredentialRepository) GetUserRole(userID, schoolID uuid.UUID) (*model.UserRole, error) {
	var ur model.UserRole
	err := r.db.Where("user_id = ? AND school_id = ?", userID, schoolID).First(&ur).Error
	return &ur, err
}

func (r *CredentialRepository) ListUserRoles(userID uuid.UUID) ([]model.UserRole, error) {
	var rows []model.UserRole
	err := r.db.Where("user_id = ?", userID).Order("created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *CredentialRepository) ListUserRolesForSchool(schoolID uuid.UUID) ([]model.UserRole, error) {
	var rows []model.UserRole
	err := r.db.Where("school_id = ?", schoolID).Order("created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *CredentialRepository) UpdateUserRole(userID, schoolID, roleID uuid.UUID) error {
	return r.db.Model(&model.UserRole{}).
		Where("user_id = ? AND school_id = ?", userID, schoolID).
		Update("role_id", roleID).Error
}

func (r *CredentialRepository) RemoveUserRole(userID, schoolID uuid.UUID) error {
	return r.db.Where("user_id = ? AND school_id = ?", userID, schoolID).Delete(&model.UserRole{}).Error
}

func (r *CredentialRepository) DeleteAllUserRoles(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error
}

func (r *CredentialRepository) ListUserIDsForSchoolByRole(schoolID uuid.UUID, roleID *uuid.UUID) ([]uuid.UUID, error) {
	q := r.db.Model(&model.UserRole{}).Where("school_id = ?", schoolID)
	if roleID != nil {
		q = q.Where("role_id = ?", *roleID)
	}
	var ids []uuid.UUID
	err := q.Pluck("user_id", &ids).Error
	return ids, err
}
