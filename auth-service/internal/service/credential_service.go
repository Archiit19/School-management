package service

import (
	"errors"
	"fmt"

	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/Archiit19/School-management/auth-service/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CredentialService struct {
	repo *repository.CredentialRepository
	rbac *RBACService
}

func NewCredentialService(repo *repository.CredentialRepository, rbac *RBACService) *CredentialService {
	return &CredentialService{repo: repo, rbac: rbac}
}

func (s *CredentialService) SetPassword(userID uuid.UUID, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return s.repo.SetCredential(&model.UserCredential{UserID: userID, PasswordHash: string(hash)})
}

func (s *CredentialService) VerifyPassword(userID uuid.UUID, password string) error {
	cred, err := s.repo.GetCredentialByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid email or password")
		}
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(password)); err != nil {
		return errors.New("invalid email or password")
	}
	return nil
}

func (s *CredentialService) DeleteCredential(userID uuid.UUID) error {
	return s.repo.DeleteCredential(userID)
}

func (s *CredentialService) AssignUserRole(userID, schoolID, roleID uuid.UUID) error {
	if _, err := s.rbac.GetRoleByID(roleID); err != nil {
		return errors.New("role not found")
	}
	if _, err := s.repo.GetUserRole(userID, schoolID); err == nil {
		return errors.New("user already has a role for this school")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.repo.AssignUserRole(&model.UserRole{UserID: userID, SchoolID: schoolID, RoleID: roleID})
}

func (s *CredentialService) UpdateUserRole(userID, schoolID, roleID uuid.UUID) error {
	if _, err := s.rbac.GetRoleByID(roleID); err != nil {
		return errors.New("role not found")
	}
	if _, err := s.repo.GetUserRole(userID, schoolID); err != nil {
		return errors.New("user role not found")
	}
	return s.repo.UpdateUserRole(userID, schoolID, roleID)
}

func (s *CredentialService) RemoveUserRole(userID, schoolID uuid.UUID) error {
	return s.repo.RemoveUserRole(userID, schoolID)
}

func (s *CredentialService) RemoveUserCompletely(userID uuid.UUID) error {
	if err := s.repo.DeleteAllUserRoles(userID); err != nil {
		return err
	}
	return s.repo.DeleteCredential(userID)
}

func (s *CredentialService) GetUserRole(userID, schoolID uuid.UUID) (*model.UserRole, error) {
	return s.repo.GetUserRole(userID, schoolID)
}

func (s *CredentialService) ListUserRoles(userID uuid.UUID) ([]model.UserRole, error) {
	return s.repo.ListUserRoles(userID)
}

func (s *CredentialService) ListUserRolesForSchool(schoolID uuid.UUID) ([]model.UserRole, error) {
	return s.repo.ListUserRolesForSchool(schoolID)
}

func (s *CredentialService) StudentRoleID(schoolID uuid.UUID) (uuid.UUID, error) {
	role, err := s.rbac.GetRoleByNameAndSchool("student", schoolID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("student role not found for school: %w", err)
	}
	return role.ID, nil
}
