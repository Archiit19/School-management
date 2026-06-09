package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Archiit19/School-management/user-service/internal/config"
	"github.com/Archiit19/School-management/user-service/internal/model"
	"github.com/Archiit19/School-management/user-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	repo   *repository.UserRepository
	auth   *authClient
	school *schoolClient
}

func NewUserService(repo *repository.UserRepository, cfg *config.Config) *UserService {
	return &UserService{
		repo:   repo,
		auth:   newAuthClient(cfg),
		school: newSchoolClient(cfg),
	}
}

func (s *UserService) rollbackCreate(userID uuid.UUID, schoolID *uuid.UUID) {
	if schoolID != nil {
		_ = s.auth.RemoveUserRole(userID, *schoolID)
		_ = s.school.RemoveMember(*schoolID, userID)
	}
	_ = s.auth.DeleteUserAuth(userID)
	_ = s.repo.Delete(userID)
}

func (s *UserService) CreateProfileInternal(req model.CreateProfileInternalRequest) (*model.User, error) {
	if _, err := s.repo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("user with this email already exists")
	}
	var studentUUID *uuid.UUID
	if req.StudentID != "" {
		sid, err := uuid.Parse(req.StudentID)
		if err != nil {
			return nil, errors.New("invalid student_id")
		}
		studentUUID = &sid
	}
	user := &model.User{Name: req.Name, Email: req.Email, StudentID: studentUUID, IsActive: true}
	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (s *UserService) CreateUser(req model.CreateUserRequest, schoolID uuid.UUID) (*model.User, error) {
	if _, err := s.repo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("user with this email already exists")
	}
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, errors.New("invalid role_id format")
	}
	roleName, err := s.auth.GetRoleByID(roleID)
	if err != nil || roleName == "" {
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

	user := &model.User{Name: req.Name, Email: req.Email, StudentID: studentUUID, IsActive: true}
	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.auth.SetCredential(user.ID, req.Password); err != nil {
		s.rollbackCreate(user.ID, nil)
		return nil, fmt.Errorf("failed to set credentials: %w", err)
	}
	if err := s.school.AddMember(schoolID, user.ID); err != nil {
		s.rollbackCreate(user.ID, &schoolID)
		return nil, fmt.Errorf("failed to link user to school: %w", err)
	}
	if err := s.auth.AssignUserRole(user.ID, schoolID, roleID); err != nil {
		s.rollbackCreate(user.ID, &schoolID)
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}

	user.SchoolID = &schoolID
	user.RoleID = &roleID
	user.RoleName = roleName
	return user, nil
}

func (s *UserService) GetUsers(schoolID uuid.UUID, query model.UserListQuery) (*model.UserListResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	memberIDs, err := s.school.ListMemberUserIDs(schoolID)
	if err != nil {
		return nil, fmt.Errorf("failed to list school members: %w", err)
	}

	if query.RoleID != "" {
		rid, err := uuid.Parse(query.RoleID)
		if err == nil {
			filtered := make([]uuid.UUID, 0)
			for _, uid := range memberIDs {
				ur, err := s.auth.GetUserRole(uid, schoolID)
				if err == nil && ur.RoleID == rid {
					filtered = append(filtered, uid)
				}
			}
			memberIDs = filtered
		}
	}

	users, total, err := s.repo.GetByIDs(memberIDs, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	for i := range users {
		if ur, err := s.auth.GetUserRole(users[i].ID, schoolID); err == nil {
			users[i].RoleID = &ur.RoleID
			users[i].RoleName = ur.RoleName
		}
		users[i].SchoolID = &schoolID
	}

	return &model.UserListResponse{Users: users, Total: total, Page: query.Page, Limit: query.Limit}, nil
}

func (s *UserService) GetUserByID(id uuid.UUID, schoolID uuid.UUID) (*model.User, error) {
	if err := s.school.GetMembership(schoolID, id); err != nil {
		return nil, errors.New("user not found")
	}
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	if ur, err := s.auth.GetUserRole(id, schoolID); err == nil {
		user.RoleID = &ur.RoleID
		user.RoleName = ur.RoleName
	}
	user.SchoolID = &schoolID
	return user, nil
}

func (s *UserService) GetUserForInternal(id uuid.UUID, schoolID *uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	if schoolID != nil && *schoolID != uuid.Nil {
		if ur, err := s.auth.GetUserRole(id, *schoolID); err == nil {
			user.SchoolID = schoolID
			user.RoleID = &ur.RoleID
			user.RoleName = ur.RoleName
		}
	} else {
		schools, err := s.school.ListMembershipsForUser(id)
		if err == nil && len(schools) == 1 {
			sid := schools[0]
			if ur, err := s.auth.GetUserRole(id, sid); err == nil {
				user.SchoolID = &sid
				user.RoleID = &ur.RoleID
				user.RoleName = ur.RoleName
			}
		}
	}
	return user, nil
}

func (s *UserService) UpdateUser(id uuid.UUID, req model.UpdateUserRequest, schoolID uuid.UUID) (*model.User, error) {
	if err := s.school.GetMembership(schoolID, id); err != nil {
		return nil, errors.New("user not found")
	}
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		existing, err := s.repo.GetByEmail(*req.Email)
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
		roleName, err := s.auth.GetRoleByID(roleID)
		if err != nil || roleName == "" {
			return nil, errors.New("role not found")
		}
		if strings.EqualFold(roleName, "student") {
			hasNew := req.StudentID != nil && strings.TrimSpace(*req.StudentID) != ""
			if user.StudentID == nil && !hasNew {
				return nil, errors.New("student role requires student_id")
			}
		}
		if err := s.auth.UpdateUserRole(id, schoolID, roleID); err != nil {
			return nil, fmt.Errorf("failed to update role: %w", err)
		}
		user.RoleID = &roleID
		user.RoleName = roleName
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
	if err := s.repo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	user.SchoolID = &schoolID
	return user, nil
}

func (s *UserService) CreateStudentLogin(req model.CreateStudentLoginRequest) (*model.User, error) {
	schoolID, _ := uuid.Parse(req.SchoolID)
	studentID, _ := uuid.Parse(req.StudentID)
	if _, err := s.repo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("user with this email already exists")
	}
	roleID, err := s.auth.StudentRoleID(schoolID)
	if err != nil {
		return nil, fmt.Errorf("could not resolve student role: %w", err)
	}
	user := &model.User{Name: req.Name, Email: req.Email, StudentID: &studentID, IsActive: true}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	if err := s.auth.SetCredential(user.ID, req.Password); err != nil {
		s.rollbackCreate(user.ID, nil)
		return nil, err
	}
	if err := s.school.AddMember(schoolID, user.ID); err != nil {
		s.rollbackCreate(user.ID, &schoolID)
		return nil, err
	}
	if err := s.auth.AssignUserRole(user.ID, schoolID, roleID); err != nil {
		s.rollbackCreate(user.ID, &schoolID)
		return nil, err
	}
	user.SchoolID = &schoolID
	user.RoleID = &roleID
	user.RoleName = "student"
	return user, nil
}

func (s *UserService) DeleteUser(id uuid.UUID, schoolID uuid.UUID, requestingUserID uuid.UUID) error {
	if id == requestingUserID {
		return errors.New("you cannot delete your own account")
	}
	if err := s.school.GetMembership(schoolID, id); err != nil {
		return errors.New("user not found")
	}
	if err := s.auth.RemoveUserRole(id, schoolID); err != nil {
		return err
	}
	if err := s.school.RemoveMember(schoolID, id); err != nil {
		return err
	}
	schools, err := s.school.ListMembershipsForUser(id)
	if err != nil {
		return err
	}
	if len(schools) == 0 {
		if err := s.auth.DeleteUserAuth(id); err != nil {
			return err
		}
		if err := s.repo.Delete(id); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
	}
	return nil
}

func (s *UserService) GetUserForInternalByEmail(email string) (*model.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) DeleteProfileInternal(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *UserService) UpdateProfileInternal(id uuid.UUID, name, email *string) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if name != nil {
		user.Name = *name
	}
	if email != nil && *email != user.Email {
		if _, err := s.repo.GetByEmail(*email); err == nil {
			return nil, errors.New("email already in use")
		}
		user.Email = *email
	}
	if err := s.repo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}
