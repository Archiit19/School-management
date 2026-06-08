package service

import (
	"errors"
	"fmt"

	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/Archiit19/School-management/auth-service/internal/repository"
	"github.com/google/uuid"
)

type RBACService struct {
	repo *repository.RBACRepository
}

func NewRBACService(repo *repository.RBACRepository) *RBACService {
	return &RBACService{repo: repo}
}

func (s *RBACService) CreateRole(req model.CreateRoleRequest, schoolID uuid.UUID) (*model.Role, error) {
	if _, err := s.repo.GetRoleByNameAndSchool(req.Name, schoolID); err == nil {
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

func (s *RBACService) CreateRoleInternal(req model.CreateRoleRequest) (*model.Role, error) {
	schoolID, err := uuid.Parse(req.SchoolID)
	if err != nil {
		return nil, errors.New("invalid school_id")
	}
	return s.CreateRole(req, schoolID)
}

func (s *RBACService) GetRoleByID(id uuid.UUID) (*model.Role, error) {
	return s.repo.GetRoleByID(id)
}

func (s *RBACService) GetRolesBySchoolID(schoolID uuid.UUID) ([]model.Role, error) {
	return s.repo.GetRolesBySchoolID(schoolID)
}

func (s *RBACService) GetRoleByNameAndSchool(name string, schoolID uuid.UUID) (*model.Role, error) {
	return s.repo.GetRoleByNameAndSchool(name, schoolID)
}

func (s *RBACService) CreatePermission(req model.CreatePermissionRequest) (*model.Permission, error) {
	if _, err := s.repo.GetPermissionByName(req.Name); err == nil {
		return nil, errors.New("permission with this name already exists")
	}
	perm := &model.Permission{Name: req.Name, Description: req.Description}
	if err := s.repo.CreatePermission(perm); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}
	return perm, nil
}

func (s *RBACService) GetAllPermissions() ([]model.Permission, error) {
	return s.repo.GetAllPermissions()
}

func (s *RBACService) AssignPermissionToRole(req model.AssignPermissionRequest) (*model.RolePermission, error) {
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, errors.New("invalid role_id")
	}
	permID, err := uuid.Parse(req.PermissionID)
	if err != nil {
		return nil, errors.New("invalid permission_id")
	}
	if _, err := s.repo.GetRoleByID(roleID); err != nil {
		return nil, errors.New("role not found")
	}
	if _, err := s.repo.GetPermissionByID(permID); err != nil {
		return nil, errors.New("permission not found")
	}
	rp := &model.RolePermission{RoleID: roleID, PermissionID: permID}
	if err := s.repo.AssignPermissionToRole(rp); err != nil {
		return nil, fmt.Errorf("failed to assign permission: %w", err)
	}
	return rp, nil
}

func (s *RBACService) GetPermissionsByRoleID(roleID uuid.UUID) ([]model.Permission, error) {
	return s.repo.GetPermissionsByRoleID(roleID)
}

func (s *RBACService) RemovePermissionFromRole(roleID, permissionID uuid.UUID) error {
	if _, err := s.repo.GetRoleByID(roleID); err != nil {
		return errors.New("role not found")
	}
	if _, err := s.repo.GetPermissionByID(permissionID); err != nil {
		return errors.New("permission not found")
	}
	return s.repo.RemovePermissionFromRole(roleID, permissionID)
}

func (s *RBACService) RoleName(roleID uuid.UUID) string {
	role, err := s.repo.GetRoleByID(roleID)
	if err != nil {
		return ""
	}
	return role.Name
}

func (s *RBACService) RolePermissionNames(roleID uuid.UUID) []string {
	perms, err := s.repo.GetPermissionsByRoleID(roleID)
	if err != nil {
		return nil
	}
	names := make([]string, len(perms))
	for i, p := range perms {
		names[i] = p.Name
	}
	return names
}
