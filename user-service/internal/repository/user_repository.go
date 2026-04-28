package repository

import (
	"github.com/avaneeshravat/school-management/user-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// ─── Roles ──────────────────────────────────────────────────────────

func (r *UserRepository) CreateRole(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *UserRepository) GetRoleByID(id uuid.UUID) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("id = ?", id).First(&role).Error
	return &role, err
}

func (r *UserRepository) GetRolesBySchoolID(schoolID uuid.UUID) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Where("school_id = ?", schoolID).Find(&roles).Error
	return roles, err
}

// ListDistinctSchoolIDsFromRoles returns every school_id that has at least one role (for backfill sync).
func (r *UserRepository) ListDistinctSchoolIDsFromRoles() ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&model.Role{}).Distinct("school_id").Pluck("school_id", &ids).Error
	return ids, err
}

func (r *UserRepository) GetRoleByNameAndSchool(name string, schoolID uuid.UUID) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("name = ? AND school_id = ?", name, schoolID).First(&role).Error
	return &role, err
}

// ─── Permissions ────────────────────────────────────────────────────

func (r *UserRepository) CreatePermission(perm *model.Permission) error {
	return r.db.Create(perm).Error
}

func (r *UserRepository) GetPermissionByID(id uuid.UUID) (*model.Permission, error) {
	var perm model.Permission
	err := r.db.Where("id = ?", id).First(&perm).Error
	return &perm, err
}

func (r *UserRepository) GetAllPermissions() ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.Find(&perms).Error
	return perms, err
}

func (r *UserRepository) GetPermissionByName(name string) (*model.Permission, error) {
	var perm model.Permission
	err := r.db.Where("name = ?", name).First(&perm).Error
	return &perm, err
}

// ─── Role-Permission Mapping ────────────────────────────────────────

func (r *UserRepository) AssignPermissionToRole(rp *model.RolePermission) error {
	return r.db.Create(rp).Error
}

// AssignPermissionToRoleIfMissing inserts the mapping only if it does not already exist.
func (r *UserRepository) AssignPermissionToRoleIfMissing(roleID, permissionID uuid.UUID) error {
	var n int64
	r.db.Model(&model.RolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&n)
	if n > 0 {
		return nil
	}
	rp := &model.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}
	return r.db.Create(rp).Error
}

func (r *UserRepository) GetPermissionsByRoleID(roleID uuid.UUID) ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&perms).Error
	return perms, err
}
