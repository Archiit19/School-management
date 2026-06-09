package service

import (
	"errors"
	"fmt"
	"log"

	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/Archiit19/School-management/auth-service/internal/rbacdata"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *RBACService) BootstrapSchoolRoles(schoolID uuid.UUID) (uuid.UUID, error) {
	templates, err := rbacdata.LoadRoleTemplates()
	if err != nil {
		return uuid.Nil, err
	}

	allPerms, err := s.repo.GetAllPermissions()
	if err != nil {
		return uuid.Nil, err
	}

	permByName := make(map[string]uuid.UUID, len(allPerms))
	for _, p := range allPerms {
		permByName[p.Name] = p.ID
	}

	for _, t := range templates {
		existing, err := s.repo.GetRoleByNameAndSchool(t.Name, schoolID)
		var role *model.Role
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return uuid.Nil, fmt.Errorf("lookup role %q: %w", t.Name, err)
			}
			nr := &model.Role{SchoolID: schoolID, Name: t.Name, Description: t.Description}
			if err := s.repo.CreateRole(nr); err != nil {
				return uuid.Nil, fmt.Errorf("create role %q: %w", t.Name, err)
			}
			role = nr
		} else {
			role = existing
		}

		var permIDs []uuid.UUID
		if len(t.Permissions) == 1 && t.Permissions[0] == "*" {
			for _, p := range allPerms {
				permIDs = append(permIDs, p.ID)
			}
		} else {
			for _, pname := range t.Permissions {
				pid, ok := permByName[pname]
				if !ok {
					log.Printf("rbac bootstrap: unknown permission %q for role %q — skipping", pname, t.Name)
					continue
				}
				permIDs = append(permIDs, pid)
			}
		}

		for _, pid := range permIDs {
			if err := s.repo.AssignPermissionToRoleIfMissing(role.ID, pid); err != nil {
				return uuid.Nil, fmt.Errorf("assign permission to role %q: %w", t.Name, err)
			}
		}

		if err := s.bootstrapRoleFields(role.ID, t.Name); err != nil {
			log.Printf("rbac bootstrap: role fields for %q: %v", t.Name, err)
		}
	}

	admin, err := s.repo.GetRoleByNameAndSchool("super_admin", schoolID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("super_admin role missing after bootstrap: %w", err)
	}
	return admin.ID, nil
}

func (s *RBACService) SyncTemplateRolesForAllSchools() error {
	ids, err := s.repo.ListDistinctSchoolIDsFromRoles()
	if err != nil {
		return err
	}
	for _, sid := range ids {
		if _, err := s.BootstrapSchoolRoles(sid); err != nil {
			return fmt.Errorf("school %s: %w", sid.String(), err)
		}
	}
	return nil
}

func (s *RBACService) bootstrapRoleFields(roleID uuid.UUID, roleName string) error {
	templates, err := rbacdata.LoadRoleFieldTemplates()
	if err != nil {
		return err
	}
	defs, ok := templates[roleName]
	if !ok || len(defs) == 0 {
		return nil
	}
	if _, err := s.repo.GetRoleFields(roleID); err == nil {
		return nil
	}
	fields := make([]model.FieldDefinition, len(defs))
	for i, d := range defs {
		fields[i] = model.FieldDefinition{
			Key:      d.Key,
			Label:    d.Label,
			Type:     d.Type,
			Required: d.Required,
			Options:  d.Options,
		}
	}
	return s.saveRoleFields(roleID, fields)
}
