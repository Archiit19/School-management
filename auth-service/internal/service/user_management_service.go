package service

import (
	"errors"
	"fmt"

	"github.com/avaneeshravat/school-management/auth-service/internal/config"
	"github.com/avaneeshravat/school-management/auth-service/internal/model"
	"github.com/avaneeshravat/school-management/auth-service/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserManagementService struct {
	repo *repository.AuthRepository
	cfg  *config.Config
	auth *AuthService // reuse for inter-service helpers
}

func NewUserManagementService(repo *repository.AuthRepository, cfg *config.Config, auth *AuthService) *UserManagementService {
	return &UserManagementService{repo: repo, cfg: cfg, auth: auth}
}

// ─── Create User ────────────────────────────────────────────────────

func (s *UserManagementService) CreateUser(req model.CreateUserRequest, schoolID uuid.UUID) (*model.User, error) {
	// 1. Check if email already exists
	_, err := s.repo.GetUserByEmail(req.Email)
	if err == nil {
		return nil, errors.New("user with this email already exists")
	}

	// 2. Validate role exists (via user-service)
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, errors.New("invalid role_id format")
	}

	roleName := s.auth.fetchRoleName(roleID)
	if roleName == "" {
		return nil, errors.New("role not found — make sure the role_id is valid")
	}

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Create user
	user := &model.User{
		SchoolID: schoolID,
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		RoleID:   roleID,
		IsActive: true,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.RoleName = roleName
	return user, nil
}

// ─── List Users ─────────────────────────────────────────────────────

func (s *UserManagementService) GetUsers(schoolID uuid.UUID, query model.UserListQuery) (*model.UserListResponse, error) {
	// Ensure sane defaults
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	users, total, err := s.repo.GetUsersBySchoolID(schoolID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// Enrich each user with role name
	for i := range users {
		users[i].RoleName = s.auth.fetchRoleName(users[i].RoleID)
	}

	return &model.UserListResponse{
		Users: users,
		Total: total,
		Page:  query.Page,
		Limit: query.Limit,
	}, nil
}

// ─── Get Single User ────────────────────────────────────────────────

func (s *UserManagementService) GetUserByID(id uuid.UUID, schoolID uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetUserByIDAndSchool(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	user.RoleName = s.auth.fetchRoleName(user.RoleID)
	return user, nil
}

// ─── Update User ────────────────────────────────────────────────────

func (s *UserManagementService) UpdateUser(id uuid.UUID, req model.UpdateUserRequest, schoolID uuid.UUID) (*model.User, error) {
	// 1. Fetch existing user (scoped to school)
	user, err := s.repo.GetUserByIDAndSchool(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// 2. Apply partial updates
	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Email != nil {
		// Check email uniqueness
		existing, err := s.repo.GetUserByEmail(*req.Email)
		if err == nil && existing.ID != user.ID {
			return nil, errors.New("email already in use by another user")
		}
		user.Email = *req.Email
	}

	if req.RoleID != nil {
		roleID, err := uuid.Parse(*req.RoleID)
		if err != nil {
			return nil, errors.New("invalid role_id format")
		}
		// Validate role exists
		roleName := s.auth.fetchRoleName(roleID)
		if roleName == "" {
			return nil, errors.New("role not found — make sure the role_id is valid")
		}
		user.RoleID = roleID
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// 3. Save
	if err := s.repo.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	user.RoleName = s.auth.fetchRoleName(user.RoleID)
	return user, nil
}

// ─── Delete User (hard delete) ──────────────────────────────────────

func (s *UserManagementService) DeleteUser(id uuid.UUID, schoolID uuid.UUID, requestingUserID uuid.UUID) error {
	// Prevent self-deletion
	if id == requestingUserID {
		return errors.New("you cannot delete your own account")
	}

	// Check user exists in this school
	_, err := s.repo.GetUserByIDAndSchool(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	if err := s.repo.DeleteUser(id, schoolID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
