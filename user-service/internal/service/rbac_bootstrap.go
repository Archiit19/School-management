package service

import (
	"errors"
	"fmt"
	"log"

	"github.com/avaneeshravat/school-management/user-service/internal/model"
	"github.com/avaneeshravat/school-management/user-service/internal/rbacdata"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BootstrapSchoolRoles creates or updates template roles for a school and attaches permissions from role_templates.json.
// Returns the super_admin role ID for linking the first admin user at registration.
func (s *UserService) BootstrapSchoolRoles(schoolID uuid.UUID) (uuid.UUID, error) {
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
			nr := &model.Role{
				SchoolID:    schoolID,
				Name:        t.Name,
				Description: t.Description,
			}
			if err := s.repo.CreateRole(nr); err != nil {
				return uuid.Nil, fmt.Errorf("create role %q: %w", t.Name, err)
			}
			role = nr
		} else {
			role = existing
		}

		var permIDs []uuid.UUID
		if len(t.Permissions) == 1 && t.Permissions[0] == "*" {
			permIDs = make([]uuid.UUID, 0, len(allPerms))
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
	}

	admin, err := s.repo.GetRoleByNameAndSchool("super_admin", schoolID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("super_admin role missing after bootstrap: %w", err)
	}
	return admin.ID, nil
}

// SyncTemplateRolesForAllSchools applies role_templates.json to every school that already has roles (startup backfill).
func (s *UserService) SyncTemplateRolesForAllSchools() error {
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
