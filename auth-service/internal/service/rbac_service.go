package service

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/Archiit19/School-management/pkg/logger"
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
		log.Error("create role: database insert failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("name", req.Name))
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	if len(req.Fields) > 0 {
		if err := s.saveRoleFields(role.ID, req.Fields); err != nil {
			return nil, err
		}
	}
	log.Info("role created", log.AddField("role_id", role.ID), log.AddField("school_id", schoolID), log.AddField("name", role.Name))
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
	roles, err := s.repo.GetRolesBySchoolID(schoolID)
	if err != nil {
		log.Error("list roles: database query failed", log.Err(err), log.AddField("school_id", schoolID))
		return nil, err
	}
	return roles, nil
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
		log.Error("create permission: database insert failed", log.Err(err), log.AddField("name", req.Name))
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}
	log.Info("permission created", log.AddField("permission_id", perm.ID), log.AddField("name", perm.Name))
	return perm, nil
}

func (s *RBACService) GetAllPermissions() ([]model.Permission, error) {
	perms, err := s.repo.GetAllPermissions()
	if err != nil {
		log.Error("list permissions: database query failed", log.Err(err))
		return nil, err
	}
	return perms, nil
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
		log.Error("assign permission to role: database insert failed", log.Err(err), log.AddField("role_id", roleID), log.AddField("permission_id", permID))
		return nil, fmt.Errorf("failed to assign permission: %w", err)
	}
	log.Info("permission assigned to role", log.AddField("role_id", roleID), log.AddField("permission_id", permID))
	return rp, nil
}

func (s *RBACService) GetPermissionsByRoleID(roleID uuid.UUID) ([]model.Permission, error) {
	perms, err := s.repo.GetPermissionsByRoleID(roleID)
	if err != nil {
		log.Error("get role permissions: database query failed", log.Err(err), log.AddField("role_id", roleID))
		return nil, err
	}
	return perms, nil
}

func (s *RBACService) RemovePermissionFromRole(roleID, permissionID uuid.UUID) error {
	if _, err := s.repo.GetRoleByID(roleID); err != nil {
		return errors.New("role not found")
	}
	if _, err := s.repo.GetPermissionByID(permissionID); err != nil {
		return errors.New("permission not found")
	}
	if err := s.repo.RemovePermissionFromRole(roleID, permissionID); err != nil {
		log.Error("remove permission from role: database delete failed", log.Err(err), log.AddField("role_id", roleID), log.AddField("permission_id", permissionID))
		return err
	}
	log.Info("permission removed from role", log.AddField("role_id", roleID), log.AddField("permission_id", permissionID))
	return nil
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

func (s *RBACService) saveRoleFields(roleID uuid.UUID, fields []model.FieldDefinition) error {
	raw, err := json.Marshal(fields)
	if err != nil {
		return fmt.Errorf("marshal role fields: %w", err)
	}
	if err := s.repo.UpsertRoleFields(&model.RoleField{RoleID: roleID, Fields: raw}); err != nil {
		log.Error("save role fields: database upsert failed", log.Err(err), log.AddField("role_id", roleID))
		return err
	}
	return nil
}

func (s *RBACService) GetRoleFields(roleID uuid.UUID) ([]model.FieldDefinition, error) {
	rf, err := s.repo.GetRoleFields(roleID)
	if err != nil {
		return nil, err
	}
	var fields []model.FieldDefinition
	if len(rf.Fields) == 0 {
		return fields, nil
	}
	if err := json.Unmarshal(rf.Fields, &fields); err != nil {
		return nil, fmt.Errorf("parse role fields: %w", err)
	}
	return fields, nil
}

func (s *RBACService) UpdateRoleFields(roleID uuid.UUID, fields []model.FieldDefinition) error {
	if _, err := s.repo.GetRoleByID(roleID); err != nil {
		return errors.New("role not found")
	}
	if err := s.saveRoleFields(roleID, fields); err != nil {
		return err
	}
	log.Info("role fields updated", log.AddField("role_id", roleID))
	return nil
}
