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

func (r *UserRepository) GetPermissionsByRoleID(roleID uuid.UUID) ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&perms).Error
	return perms, err
}
