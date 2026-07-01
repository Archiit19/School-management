package service

import (
	"errors"
	"fmt"

	log "github.com/Archiit19/School-management/pkg/logger"
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
		log.Error("set password: hash failed", log.Err(err), log.AddField("user_id", userID))
		return fmt.Errorf("failed to hash password: %w", err)
	}
	if err := s.repo.SetCredential(&model.UserCredential{UserID: userID, PasswordHash: string(hash)}); err != nil {
		log.Error("set password: database save failed", log.Err(err), log.AddField("user_id", userID))
		return err
	}
	log.Info("credential set", log.AddField("user_id", userID))
	return nil
}

func (s *CredentialService) VerifyPassword(userID uuid.UUID, password string) error {
	cred, err := s.repo.GetCredentialByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid email or password")
		}
		log.Error("verify password: database fetch failed", log.Err(err), log.AddField("user_id", userID))
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(password)); err != nil {
		return errors.New("invalid email or password")
	}
	return nil
}

func (s *CredentialService) DeleteCredential(userID uuid.UUID) error {
	if err := s.repo.DeleteCredential(userID); err != nil {
		log.Error("delete credential: database delete failed", log.Err(err), log.AddField("user_id", userID))
		return err
	}
	log.Info("credential deleted", log.AddField("user_id", userID))
	return nil
}

func (s *CredentialService) AssignUserRole(userID, schoolID, roleID uuid.UUID) error {
	if _, err := s.rbac.GetRoleByID(roleID); err != nil {
		return errors.New("role not found")
	}
	if _, err := s.repo.GetUserRole(userID, schoolID); err == nil {
		return errors.New("user already has a role for this school")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("assign user role: lookup failed", log.Err(err), log.AddField("user_id", userID), log.AddField("school_id", schoolID))
		return err
	}
	if err := s.repo.AssignUserRole(&model.UserRole{UserID: userID, SchoolID: schoolID, RoleID: roleID}); err != nil {
		log.Error("assign user role: database insert failed", log.Err(err), log.AddField("user_id", userID), log.AddField("school_id", schoolID), log.AddField("role_id", roleID))
		return err
	}
	log.Info("user role assigned", log.AddField("user_id", userID), log.AddField("school_id", schoolID), log.AddField("role_id", roleID))
	return nil
}

func (s *CredentialService) UpdateUserRole(userID, schoolID, roleID uuid.UUID) error {
	if _, err := s.rbac.GetRoleByID(roleID); err != nil {
		return errors.New("role not found")
	}
	if _, err := s.repo.GetUserRole(userID, schoolID); err != nil {
		return errors.New("user role not found")
	}
	if err := s.repo.UpdateUserRole(userID, schoolID, roleID); err != nil {
		log.Error("update user role: database update failed", log.Err(err), log.AddField("user_id", userID), log.AddField("school_id", schoolID), log.AddField("role_id", roleID))
		return err
	}
	log.Info("user role updated", log.AddField("user_id", userID), log.AddField("school_id", schoolID), log.AddField("role_id", roleID))
	return nil
}

func (s *CredentialService) RemoveUserRole(userID, schoolID uuid.UUID) error {
	if err := s.repo.RemoveUserRole(userID, schoolID); err != nil {
		log.Error("remove user role: database delete failed", log.Err(err), log.AddField("user_id", userID), log.AddField("school_id", schoolID))
		return err
	}
	log.Info("user role removed", log.AddField("user_id", userID), log.AddField("school_id", schoolID))
	return nil
}

func (s *CredentialService) RemoveUserCompletely(userID uuid.UUID) error {
	if err := s.repo.DeleteAllUserRoles(userID); err != nil {
		log.Error("remove user completely: delete roles failed", log.Err(err), log.AddField("user_id", userID))
		return err
	}
	if err := s.repo.DeleteCredential(userID); err != nil {
		log.Error("remove user completely: delete credential failed", log.Err(err), log.AddField("user_id", userID))
		return err
	}
	log.Info("user credentials and roles removed", log.AddField("user_id", userID))
	return nil
}

func (s *CredentialService) GetUserRole(userID, schoolID uuid.UUID) (*model.UserRole, error) {
	return s.repo.GetUserRole(userID, schoolID)
}

func (s *CredentialService) ListUserRoles(userID uuid.UUID) ([]model.UserRole, error) {
	rows, err := s.repo.ListUserRoles(userID)
	if err != nil {
		log.Error("list user roles: database query failed", log.Err(err), log.AddField("user_id", userID))
		return nil, err
	}
	return rows, nil
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
