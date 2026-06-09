package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Archiit19/School-management/auth-service/internal/config"
	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthService struct {
	cfg    *config.Config
	school *schoolClient
	users  *userClient
	creds  *CredentialService
	rbac   *RBACService
}

func NewAuthService(cfg *config.Config, creds *CredentialService, rbac *RBACService) *AuthService {
	return &AuthService{
		cfg:    cfg,
		school: newSchoolClient(cfg),
		users:  newUserClient(cfg),
		creds:  creds,
		rbac:   rbac,
	}
}

type tokenContext struct {
	SchoolID    *uuid.UUID
	RoleID      *uuid.UUID
	RoleName    string
	Permissions []string
}

func (s *AuthService) Signup(req model.SignupRequest) (*model.LoginResponse, error) {
	if _, err := s.users.GetByEmail(req.Email); err == nil {
		return nil, errors.New("user with this email already exists")
	} else if !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}

	user, err := s.users.CreateProfile(req.Name, req.Email, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user profile: %w", err)
	}
	if err := s.creds.SetPassword(user.ID, req.Password); err != nil {
		_ = s.users.DeleteProfile(user.ID)
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	user.RoleName = platformAdminRole
	token, err := s.generateToken(user, tokenContext{
		RoleName:    platformAdminRole,
		Permissions: platformAdminPermissions,
	})
	if err != nil {
		return nil, err
	}
	return &model.LoginResponse{Token: token, User: *user}, nil
}

func (s *AuthService) RegisterSchool(req model.RegisterSchoolRequest) (*model.RegisterSchoolResponse, error) {
	if existing, err := s.school.GetSchoolByEmail(req.SchoolEmail); err == nil && existing != nil {
		return nil, errors.New("school with this email already exists")
	}
	if _, err := s.users.GetByEmail(req.AdminEmail); err == nil {
		return nil, errors.New("user with this email already exists")
	}

	user, err := s.users.CreateProfile(req.AdminName, req.AdminEmail, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}
	if err := s.creds.SetPassword(user.ID, req.AdminPassword); err != nil {
		_ = s.users.DeleteProfile(user.ID)
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	school, err := s.school.CreateSchoolForUser(user.ID, req.SchoolName, req.SchoolAddress, req.SchoolPhone, req.SchoolEmail)
	if err != nil {
		_ = s.creds.RemoveUserCompletely(user.ID)
		_ = s.users.DeleteProfile(user.ID)
		return nil, err
	}

	ur, err := s.creds.GetUserRole(user.ID, school.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load school role: %w", err)
	}

	roleName := s.rbac.RoleName(ur.RoleID)
	perms := s.rbac.RolePermissionNames(ur.RoleID)
	token, err := s.generateToken(user, tokenContext{
		SchoolID:    &school.ID,
		RoleID:      &ur.RoleID,
		RoleName:    roleName,
		Permissions: perms,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	user.RoleName = roleName
	user.SchoolID = &school.ID
	user.RoleID = &ur.RoleID
	return &model.RegisterSchoolResponse{School: *school, Admin: *user, Token: token}, nil
}

func (s *AuthService) hasSuperAdminRole(userID uuid.UUID, memberships []schoolMembership) bool {
	for _, m := range memberships {
		if ur, err := s.creds.GetUserRole(userID, m.SchoolID); err == nil {
			if s.rbac.RoleName(ur.RoleID) == "super_admin" {
				return true
			}
		}
	}
	return false
}

func (s *AuthService) platformLoginContext(user *model.User) tokenContext {
	user.RoleName = platformAdminRole
	user.SchoolID = nil
	user.RoleID = nil
	if schools, err := s.school.ListSchoolsForUser(user.ID); err == nil {
		user.Schools = schools
	}
	return tokenContext{RoleName: platformAdminRole, Permissions: platformAdminPermissions}
}

func (s *AuthService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.users.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}
	if err := s.creds.VerifyPassword(user.ID, req.Password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	memberships, err := s.school.ListMembershipsForUser(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load school memberships: %w", err)
	}

	var ctx tokenContext
	if len(memberships) == 0 || s.hasSuperAdminRole(user.ID, memberships) {
		ctx = s.platformLoginContext(user)
	} else if len(memberships) == 1 {
		m := memberships[0]
		ur, err := s.creds.GetUserRole(user.ID, m.SchoolID)
		if err != nil {
			return nil, fmt.Errorf("failed to load role for school: %w", err)
		}
		roleName := s.rbac.RoleName(ur.RoleID)
		user.RoleName = roleName
		user.SchoolID = &m.SchoolID
		user.RoleID = &ur.RoleID
		ctx = tokenContext{
			SchoolID:    &m.SchoolID,
			RoleID:      &ur.RoleID,
			RoleName:    roleName,
			Permissions: s.rbac.RolePermissionNames(ur.RoleID),
		}
	} else {
		ctx = s.platformLoginContext(user)
	}

	token, err := s.generateToken(user, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	return &model.LoginResponse{Token: token, User: *user}, nil
}

func (s *AuthService) SelectSchool(userID uuid.UUID, schoolID uuid.UUID) (*model.LoginResponse, error) {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	memberships, err := s.school.ListMembershipsForUser(userID)
	if err != nil {
		return nil, err
	}
	if len(memberships) == 1 {
		if ur, err := s.creds.GetUserRole(userID, memberships[0].SchoolID); err == nil {
			if s.rbac.RoleName(ur.RoleID) != "super_admin" {
				return nil, errors.New("school staff accounts cannot switch school context")
			}
		}
	}

	if _, err := s.school.GetMembership(schoolID, userID); err != nil {
		return nil, errors.New("you are not a member of this school")
	}

	ur, err := s.creds.GetUserRole(userID, schoolID)
	if err != nil {
		return nil, errors.New("no role assigned for this school")
	}

	roleName := s.rbac.RoleName(ur.RoleID)
	perms := s.rbac.RolePermissionNames(ur.RoleID)
	user.RoleName = roleName
	user.SchoolID = &schoolID
	user.RoleID = &ur.RoleID
	if school, err := s.school.GetSchoolByID(schoolID); err == nil {
		user.School = school
	}

	token, err := s.generateToken(user, tokenContext{
		SchoolID: &schoolID, RoleID: &ur.RoleID, RoleName: roleName, Permissions: perms,
	})
	if err != nil {
		return nil, err
	}
	return &model.LoginResponse{Token: token, User: *user}, nil
}

func (s *AuthService) ExitSchool(userID uuid.UUID) (*model.LoginResponse, error) {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	ctx := s.platformLoginContext(user)
	token, err := s.generateToken(user, ctx)
	if err != nil {
		return nil, err
	}
	return &model.LoginResponse{Token: token, User: *user}, nil
}

func (s *AuthService) UpdateProfile(userID uuid.UUID, req model.UpdateProfileRequest) (*model.User, error) {
	user, err := s.users.UpdateProfile(userID, req)
	if err != nil {
		return nil, err
	}
	user.RoleName = platformAdminRole
	return user, nil
}

func (s *AuthService) GetMe(userID uuid.UUID, jwtSchoolID uuid.UUID, jwtRoleName string) (*model.User, error) {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if jwtSchoolID != uuid.Nil {
		user.SchoolID = &jwtSchoolID
		user.RoleName = jwtRoleName
		if ur, err := s.creds.GetUserRole(userID, jwtSchoolID); err == nil {
			user.RoleID = &ur.RoleID
		}
		if school, err := s.school.GetSchoolByID(jwtSchoolID); err == nil {
			user.School = school
		}
	} else {
		user.RoleName = platformAdminRole
		if schools, err := s.school.ListSchoolsForUser(userID); err == nil {
			user.Schools = schools
		}
	}
	return user, nil
}

func (s *AuthService) generateToken(user *model.User, ctx tokenContext) (string, error) {
	claims := jwt.MapClaims{
		"user_id":     user.ID.String(),
		"role_name":   ctx.RoleName,
		"permissions": ctx.Permissions,
		"email":       user.Email,
		"exp":         time.Now().Add(24 * time.Hour).Unix(),
		"iat":         time.Now().Unix(),
	}
	if ctx.SchoolID != nil && *ctx.SchoolID != uuid.Nil {
		claims["school_id"] = ctx.SchoolID.String()
	}
	if ctx.RoleID != nil && *ctx.RoleID != uuid.Nil {
		claims["role_id"] = ctx.RoleID.String()
	}
	if strings.EqualFold(ctx.RoleName, "student") {
		claims["student_id"] = user.ID.String()
		if profile := s.fetchUserProfile(user.ID); profile != nil {
			if classID, ok := profile["class_id"].(string); ok && classID != "" {
				claims["class_id"] = classID
			}
			if sectionID, ok := profile["section_id"].(string); ok && sectionID != "" {
				claims["section_id"] = sectionID
			}
		}
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) fetchUserProfile(userID uuid.UUID) map[string]interface{} {
	url := fmt.Sprintf("%s/internal/users/%s/profile", s.cfg.UserServiceURL, userID.String())
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	if s.cfg.InternalServiceToken != "" {
		req.Header.Set("X-Internal-Token", s.cfg.InternalServiceToken)
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil
	}
	defer resp.Body.Close()
	var profile map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil
	}
	return profile
}
