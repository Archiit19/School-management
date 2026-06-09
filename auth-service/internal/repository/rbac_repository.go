package repository

import (
	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RBACRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) *RBACRepository {
	return &RBACRepository{db: db}
}

func (r *RBACRepository) CreateRole(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *RBACRepository) GetRoleByID(id uuid.UUID) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("id = ?", id).First(&role).Error
	return &role, err
}

func (r *RBACRepository) GetRolesBySchoolID(schoolID uuid.UUID) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Where("school_id = ?", schoolID).Find(&roles).Error
	return roles, err
}

func (r *RBACRepository) ListDistinctSchoolIDsFromRoles() ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&model.Role{}).Distinct("school_id").Pluck("school_id", &ids).Error
	return ids, err
}

func (r *RBACRepository) GetRoleByNameAndSchool(name string, schoolID uuid.UUID) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("name = ? AND school_id = ?", name, schoolID).First(&role).Error
	return &role, err
}

func (r *RBACRepository) CreatePermission(perm *model.Permission) error {
	return r.db.Create(perm).Error
}

func (r *RBACRepository) GetPermissionByID(id uuid.UUID) (*model.Permission, error) {
	var perm model.Permission
	err := r.db.Where("id = ?", id).First(&perm).Error
	return &perm, err
}

func (r *RBACRepository) GetAllPermissions() ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.Find(&perms).Error
	return perms, err
}

func (r *RBACRepository) GetPermissionByName(name string) (*model.Permission, error) {
	var perm model.Permission
	err := r.db.Where("name = ?", name).First(&perm).Error
	return &perm, err
}

func (r *RBACRepository) AssignPermissionToRole(rp *model.RolePermission) error {
	return r.db.Create(rp).Error
}

func (r *RBACRepository) AssignPermissionToRoleIfMissing(roleID, permissionID uuid.UUID) error {
	var n int64
	r.db.Model(&model.RolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&n)
	if n > 0 {
		return nil
	}
	return r.db.Create(&model.RolePermission{RoleID: roleID, PermissionID: permissionID}).Error
}

func (r *RBACRepository) GetPermissionsByRoleID(roleID uuid.UUID) ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&perms).Error
	return perms, err
}

func (r *RBACRepository) RemovePermissionFromRole(roleID, permissionID uuid.UUID) error {
	return r.db.
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&model.RolePermission{}).Error
}
