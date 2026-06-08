package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Archiit19/School-management/auth-service/internal/config"
	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/Archiit19/School-management/auth-service/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserManagementService struct {
	repo *repository.AuthRepository
	cfg  *config.Config
	auth *AuthService
}

func NewUserManagementService(repo *repository.AuthRepository, cfg *config.Config, auth *AuthService) *UserManagementService {
	return &UserManagementService{repo: repo, cfg: cfg, auth: auth}
}

func (s *UserManagementService) CreateUser(req model.CreateUserRequest, schoolID uuid.UUID) (*model.User, error) {
	_, err := s.repo.GetUserByEmail(req.Email)
	if err == nil {
		return nil, errors.New("user with this email already exists")
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, errors.New("invalid role_id format")
	}

	roleName := s.auth.fetchRoleName(&roleID)
	if roleName == "" {
		return nil, errors.New("role not found — make sure the role_id is valid")
	}

	var studentUUID *uuid.UUID
	if strings.EqualFold(roleName, "student") {
		if strings.TrimSpace(req.StudentID) == "" {
			return nil, errors.New("pupil login accounts require student_id (UUID of an admitted student)")
		}
		sid, err := uuid.Parse(req.StudentID)
		if err != nil {
			return nil, errors.New("invalid student_id")
		}
		studentUUID = &sid
	} else if strings.TrimSpace(req.StudentID) != "" {
		return nil, errors.New("student_id is only allowed when role is student")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		StudentID: studentUUID,
		IsActive:  true,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.auth.school.AddMember(schoolID, user.ID, roleID); err != nil {
		return nil, fmt.Errorf("failed to link user to school: %w", err)
	}

	user.SchoolID = &schoolID
	user.RoleID = &roleID
	user.RoleName = roleName
	return user, nil
}

func (s *UserManagementService) GetUsers(schoolID uuid.UUID, query model.UserListQuery) (*model.UserListResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	var roleFilter *uuid.UUID
	if query.RoleID != "" {
		if rid, err := uuid.Parse(query.RoleID); err == nil {
			roleFilter = &rid
		}
	}

	userIDs, err := s.auth.school.ListUserIDsForSchool(schoolID, roleFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list school members: %w", err)
	}

	users, total, err := s.repo.GetUsersByIDs(userIDs, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	roleByUser := s.roleByUserForSchool(schoolID)
	for i := range users {
		if rid, ok := roleByUser[users[i].ID]; ok {
			users[i].RoleID = &rid
			users[i].RoleName = s.auth.fetchRoleName(&rid)
		}
		users[i].SchoolID = &schoolID
	}

	return &model.UserListResponse{
		Users: users,
		Total: total,
		Page:  query.Page,
		Limit: query.Limit,
	}, nil
}

func (s *UserManagementService) roleByUserForSchool(schoolID uuid.UUID) map[uuid.UUID]uuid.UUID {
	out := make(map[uuid.UUID]uuid.UUID)
	members, err := s.auth.school.ListMembersForSchool(schoolID)
	if err != nil {
		return out
	}
	for _, m := range members {
		out[m.UserID] = m.RoleID
	}
	return out
}

func (s *UserManagementService) GetUserByID(id uuid.UUID, schoolID uuid.UUID) (*model.User, error) {
	if _, err := s.auth.school.GetMembership(schoolID, id); err != nil {
		return nil, errors.New("user not found")
	}

	user, err := s.repo.GetUserByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if m, err := s.auth.school.GetMembership(schoolID, id); err == nil {
		user.RoleID = &m.RoleID
		user.RoleName = s.auth.fetchRoleName(&m.RoleID)
	}
	user.SchoolID = &schoolID
	return user, nil
}

func (s *UserManagementService) GetUserForInternalService(id uuid.UUID, schoolID *uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if schoolID != nil && *schoolID != uuid.Nil {
		if m, err := s.auth.school.GetMembership(*schoolID, id); err == nil {
			user.SchoolID = schoolID
			user.RoleID = &m.RoleID
			user.RoleName = s.auth.fetchRoleName(&m.RoleID)
		}
	} else {
		memberships, err := s.auth.school.ListMembershipsForUser(id)
		if err == nil && len(memberships) == 1 {
			m := memberships[0]
			user.SchoolID = &m.SchoolID
			user.RoleID = &m.RoleID
			user.RoleName = s.auth.fetchRoleName(&m.RoleID)
		}
	}

	return user, nil
}

func (s *UserManagementService) UpdateUser(id uuid.UUID, req model.UpdateUserRequest, schoolID uuid.UUID) (*model.User, error) {
	if _, err := s.auth.school.GetMembership(schoolID, id); err != nil {
		return nil, errors.New("user not found")
	}

	user, err := s.repo.GetUserByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Email != nil {
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
		roleName := s.auth.fetchRoleName(&roleID)
		if roleName == "" {
			return nil, errors.New("role not found — make sure the role_id is valid")
		}
		if strings.EqualFold(roleName, "student") {
			hasNew := req.StudentID != nil && strings.TrimSpace(*req.StudentID) != ""
			if user.StudentID == nil && !hasNew {
				return nil, errors.New("student role requires student_id — set student_id to an admitted pupil UUID")
			}
		} else if req.StudentID != nil && strings.TrimSpace(*req.StudentID) != "" {
			return nil, errors.New("student_id is only valid for the student role")
		}
		if err := s.auth.school.UpdateMemberRole(schoolID, id, roleID); err != nil {
			return nil, fmt.Errorf("failed to update school membership: %w", err)
		}
		user.RoleID = &roleID
	}

	if req.StudentID != nil {
		v := strings.TrimSpace(*req.StudentID)
		if v == "" {
			user.StudentID = nil
		} else {
			sid, err := uuid.Parse(v)
			if err != nil {
				return nil, errors.New("invalid student_id")
			}
			user.StudentID = &sid
		}
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if user.RoleID != nil {
		user.RoleName = s.auth.fetchRoleName(user.RoleID)
	}
	user.SchoolID = &schoolID
	return user, nil
}

func (s *UserManagementService) CreateStudentLogin(req model.CreateStudentLoginRequest) (*model.User, error) {
	schoolID, err := uuid.Parse(req.SchoolID)
	if err != nil {
		return nil, errors.New("invalid school_id")
	}
	studentID, err := uuid.Parse(req.StudentID)
	if err != nil {
		return nil, errors.New("invalid student_id")
	}

	if _, err := s.repo.GetUserByEmail(req.Email); err == nil {
		return nil, errors.New("user with this email already exists")
	}

	roleID, err := s.auth.fetchStudentRoleID(schoolID)
	if err != nil {
		return nil, fmt.Errorf("could not resolve student role for school: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		StudentID: &studentID,
		IsActive:  true,
	}
	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create student login: %w", err)
	}

	if err := s.auth.school.AddMember(schoolID, user.ID, roleID); err != nil {
		return nil, fmt.Errorf("failed to link student to school: %w", err)
	}

	user.SchoolID = &schoolID
	user.RoleID = &roleID
	user.RoleName = "student"
	return user, nil
}

func (s *UserManagementService) DeleteUser(id uuid.UUID, schoolID uuid.UUID, requestingUserID uuid.UUID) error {
	if id == requestingUserID {
		return errors.New("you cannot delete your own account")
	}

	if _, err := s.auth.school.GetMembership(schoolID, id); err != nil {
		return errors.New("user not found")
	}

	if err := s.auth.school.RemoveMember(schoolID, id); err != nil {
		return fmt.Errorf("failed to remove school membership: %w", err)
	}

	memberships, err := s.auth.school.ListMembershipsForUser(id)
	if err != nil {
		return err
	}
	if len(memberships) == 0 {
		if err := s.repo.DeleteUser(id); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
	}

	return nil
}
