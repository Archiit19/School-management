package service

import (
	"errors"
	"fmt"

	"github.com/avaneeshravat/school-management/user-service/internal/model"
	"github.com/avaneeshravat/school-management/user-service/internal/repository"
	"github.com/google/uuid"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// ─── Roles ──────────────────────────────────────────────────────────

// CreateRole creates a role scoped to a school.
func (s *UserService) CreateRole(req model.CreateRoleRequest, schoolID uuid.UUID) (*model.Role, error) {
	// Check duplicate
	_, err := s.repo.GetRoleByNameAndSchool(req.Name, schoolID)
	if err == nil {
		return nil, errors.New("role with this name already exists for this school")
	}

	role := &model.Role{
		SchoolID:    schoolID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.CreateRole(role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// CreateRoleInternal is used by other services (e.g. auth-service) to create roles.
// It takes school_id from the request body instead of JWT.
func (s *UserService) CreateRoleInternal(req model.CreateRoleRequest) (*model.Role, error) {
	schoolID, err := uuid.Parse(req.SchoolID)
	if err != nil {
		return nil, errors.New("invalid school_id")
	}

	return s.CreateRole(req, schoolID)
}

func (s *UserService) GetRoleByID(id uuid.UUID) (*model.Role, error) {
	return s.repo.GetRoleByID(id)
}

func (s *UserService) GetRolesBySchoolID(schoolID uuid.UUID) ([]model.Role, error) {
	return s.repo.GetRolesBySchoolID(schoolID)
}

// ─── Permissions ────────────────────────────────────────────────────

func (s *UserService) CreatePermission(req model.CreatePermissionRequest) (*model.Permission, error) {
	// Check duplicate
	_, err := s.repo.GetPermissionByName(req.Name)
	if err == nil {
		return nil, errors.New("permission with this name already exists")
	}

	perm := &model.Permission{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.CreatePermission(perm); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return perm, nil
}

func (s *UserService) GetAllPermissions() ([]model.Permission, error) {
	return s.repo.GetAllPermissions()
}

// ─── Role-Permission Assignment ─────────────────────────────────────

func (s *UserService) AssignPermissionToRole(req model.AssignPermissionRequest) (*model.RolePermission, error) {
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, errors.New("invalid role_id")
	}

	permID, err := uuid.Parse(req.PermissionID)
	if err != nil {
		return nil, errors.New("invalid permission_id")
	}

	// Verify role exists
	if _, err := s.repo.GetRoleByID(roleID); err != nil {
		return nil, errors.New("role not found")
	}

	// Verify permission exists
	if _, err := s.repo.GetPermissionByID(permID); err != nil {
		return nil, errors.New("permission not found")
	}

	rp := &model.RolePermission{
		RoleID:       roleID,
		PermissionID: permID,
	}

	if err := s.repo.AssignPermissionToRole(rp); err != nil {
		return nil, fmt.Errorf("failed to assign permission: %w", err)
	}

	return rp, nil
}

func (s *UserService) GetPermissionsByRoleID(roleID uuid.UUID) ([]model.Permission, error) {
	return s.repo.GetPermissionsByRoleID(roleID)
}
