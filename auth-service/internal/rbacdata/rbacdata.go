package rbacdata

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed predefined_permissions.json role_templates.json
var fs embed.FS

type PermissionEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoleTemplate struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

func LoadPredefinedPermissions() ([]PermissionEntry, error) {
	raw, err := fs.ReadFile("predefined_permissions.json")
	if err != nil {
		return nil, fmt.Errorf("read predefined_permissions.json: %w", err)
	}
	var list []PermissionEntry
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("parse predefined_permissions.json: %w", err)
	}
	return list, nil
}

func LoadRoleTemplates() ([]RoleTemplate, error) {
	raw, err := fs.ReadFile("role_templates.json")
	if err != nil {
		return nil, fmt.Errorf("read role_templates.json: %w", err)
	}
	var list []RoleTemplate
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("parse role_templates.json: %w", err)
	}
	return list, nil
}
