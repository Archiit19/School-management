package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Archiit19/School-management/auth-service/internal/config"
	"github.com/Archiit19/School-management/auth-service/internal/model"
	"github.com/Archiit19/School-management/auth-service/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	repo   *repository.AuthRepository
	cfg    *config.Config
	school *schoolClient
}

func NewAuthService(repo *repository.AuthRepository, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, cfg: cfg, school: newSchoolClient(cfg)}
}

type tokenContext struct {
	SchoolID    *uuid.UUID
	RoleID      *uuid.UUID
	RoleName    string
	Permissions []string
}

// ─── Signup (platform admin) ─────────────────────────────────────────

func (s *AuthService) Signup(req model.SignupRequest) (*model.LoginResponse, error) {
	_, err := s.repo.GetUserByEmail(req.Email)
	if err == nil {
		return nil, errors.New("user with this email already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		IsActive: true,
	}
	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
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

// ─── Register School (legacy one-step flow) ──────────────────────────

func (s *AuthService) RegisterSchool(req model.RegisterSchoolRequest) (*model.RegisterSchoolResponse, error) {
	if existing, err := s.school.GetSchoolByEmail(req.SchoolEmail); err == nil && existing != nil {
		return nil, errors.New("school with this email already exists")
	}

	_, err := s.repo.GetUserByEmail(req.AdminEmail)
	if err == nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Name:     req.AdminName,
		Email:    req.AdminEmail,
		Password: string(hashedPassword),
		IsActive: true,
	}
	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	school, err := s.school.CreateSchoolForUser(user.ID, req.SchoolName, req.SchoolAddress, req.SchoolPhone, req.SchoolEmail)
	if err != nil {
		return nil, err
	}

	m, err := s.school.GetMembership(school.ID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load school membership: %w", err)
	}

	perms := s.fetchRolePermissions(&m.RoleID)
	token, err := s.generateToken(user, tokenContext{
		SchoolID:    &school.ID,
		RoleID:      &m.RoleID,
		RoleName:    "super_admin",
		Permissions: perms,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	user.RoleName = "super_admin"
	admin := *user

	return &model.RegisterSchoolResponse{
		School: *school,
		Admin:  admin,
		Token:  token,
	}, nil
}

func (s *AuthService) hasSuperAdminMembership(memberships []schoolMembership) bool {
	for _, m := range memberships {
		if s.fetchRoleName(&m.RoleID) == "super_admin" {
			return true
		}
	}
	return false
}

// platformLoginContext is used for signup and for logins that should land on the
// platform dashboard (Dashboard + Schools), not inside a school workspace.
func (s *AuthService) platformLoginContext(user *model.User) tokenContext {
	user.RoleName = platformAdminRole
	user.SchoolID = nil
	user.RoleID = nil
	if schools, err := s.school.ListSchoolsForUser(user.ID); err == nil {
		user.Schools = schools
	}
	return tokenContext{
		RoleName:    platformAdminRole,
		Permissions: platformAdminPermissions,
	}
}

// ─── Login ──────────────────────────────────────────────────────────

func (s *AuthService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	memberships, err := s.school.ListMembershipsForUser(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load school memberships: %w", err)
	}

	var ctx tokenContext
	if len(memberships) == 0 || s.hasSuperAdminMembership(memberships) {
		// Platform hub: create schools and open/switch between them.
		ctx = s.platformLoginContext(user)
	} else if len(memberships) == 1 {
		m := memberships[0]
		roleName := s.fetchRoleName(&m.RoleID)
		user.RoleName = roleName
		user.SchoolID = &m.SchoolID
		user.RoleID = &m.RoleID
		ctx = tokenContext{
			SchoolID:    &m.SchoolID,
			RoleID:      &m.RoleID,
			RoleName:    roleName,
			Permissions: s.fetchRolePermissions(&m.RoleID),
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

// ─── Select / exit school context ───────────────────────────────────

func (s *AuthService) SelectSchool(userID uuid.UUID, schoolID uuid.UUID) (*model.LoginResponse, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	memberships, err := s.school.ListMembershipsForUser(userID)
	if err != nil {
		return nil, err
	}
	if len(memberships) == 1 {
		roleName := s.fetchRoleName(&memberships[0].RoleID)
		if roleName != "super_admin" {
			return nil, errors.New("school staff accounts cannot switch school context")
		}
	}

	m, err := s.school.GetMembership(schoolID, userID)
	if err != nil {
		return nil, errors.New("you are not a member of this school")
	}

	roleName := s.fetchRoleName(&m.RoleID)
	perms := s.fetchRolePermissions(&m.RoleID)
	user.RoleName = roleName
	user.SchoolID = &schoolID
	user.RoleID = &m.RoleID
	if school, err := s.school.GetSchoolByID(schoolID); err == nil {
		user.School = school
	}
	token, err := s.generateToken(user, tokenContext{
		SchoolID:    &schoolID,
		RoleID:      &m.RoleID,
		RoleName:    roleName,
		Permissions: perms,
	})
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{Token: token, User: *user}, nil
}

func (s *AuthService) ExitSchool(userID uuid.UUID) (*model.LoginResponse, error) {
	user, err := s.repo.GetUserByID(userID)
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

// ─── Profile ────────────────────────────────────────────────────────

func (s *AuthService) UpdateProfile(userID uuid.UUID, req model.UpdateProfileRequest) (*model.User, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil && *req.Email != user.Email {
		if _, err := s.repo.GetUserByEmail(*req.Email); err == nil {
			return nil, errors.New("email already in use")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		user.Email = *req.Email
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	user.RoleName = platformAdminRole
	return user, nil
}

// ─── Get Current User (Me) ─────────────────────────────────────────

func (s *AuthService) GetMe(userID uuid.UUID, jwtSchoolID uuid.UUID, jwtRoleName string) (*model.User, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if jwtSchoolID != uuid.Nil {
		user.SchoolID = &jwtSchoolID
		user.RoleName = jwtRoleName
		if m, err := s.school.GetMembership(jwtSchoolID, userID); err == nil {
			user.RoleID = &m.RoleID
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

// ─── JWT Helpers ────────────────────────────────────────────────────

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
	if user.StudentID != nil {
		claims["student_id"] = user.StudentID.String()
		if student := s.fetchStudentDetails(*user.StudentID); student != nil {
			claims["class_id"] = student.ClassID
			if student.SectionID != "" {
				claims["section_id"] = student.SectionID
			}
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

type studentInfo struct {
	ClassID   string `json:"class_id"`
	SectionID string `json:"section_id"`
}

func (s *AuthService) fetchStudentDetails(studentID uuid.UUID) *studentInfo {
	url := fmt.Sprintf("%s/internal/students/%s", s.cfg.StudentServiceURL, studentID.String())
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil
	}
	defer resp.Body.Close()

	var student studentInfo
	if err := json.NewDecoder(resp.Body).Decode(&student); err != nil {
		return nil
	}
	return &student
}

func (s *AuthService) bootstrapSchoolRoles(schoolID uuid.UUID) (uuid.UUID, error) {
	payload := map[string]string{"school_id": schoolID.String()}
	body, err := json.Marshal(payload)
	if err != nil {
		return uuid.Nil, err
	}
	url := fmt.Sprintf("%s/api/v1/internal/bootstrap-school", s.cfg.UserServiceURL)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return uuid.Nil, fmt.Errorf("user-service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, fmt.Errorf("user-service bootstrap returned status %d", resp.StatusCode)
	}

	var result struct {
		SuperAdminRoleID string `json:"super_admin_role_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return uuid.Nil, fmt.Errorf("failed to decode bootstrap response: %w", err)
	}

	return uuid.Parse(result.SuperAdminRoleID)
}

func (s *AuthService) fetchSuperAdminRoleID(schoolID uuid.UUID) (uuid.UUID, error) {
	url := fmt.Sprintf(
		"%s/api/v1/internal/roles/by-name?school_id=%s&name=super_admin",
		s.cfg.UserServiceURL,
		schoolID.String(),
	)
	resp, err := http.Get(url)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user-service unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return s.bootstrapSchoolRoles(schoolID)
	}
	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, fmt.Errorf("user-service returned status %d", resp.StatusCode)
	}
	var role struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(role.ID)
}

func (s *AuthService) fetchRoleName(roleID *uuid.UUID) string {
	if roleID == nil || *roleID == uuid.Nil {
		return ""
	}

	url := fmt.Sprintf("%s/api/v1/roles/%s", s.cfg.UserServiceURL, roleID.String())
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	return result.Name
}

func (s *AuthService) fetchStudentRoleID(schoolID uuid.UUID) (uuid.UUID, error) {
	url := fmt.Sprintf(
		"%s/api/v1/internal/roles/by-name?school_id=%s&name=student",
		s.cfg.UserServiceURL,
		schoolID.String(),
	)
	resp, err := http.Get(url)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user-service unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, fmt.Errorf("user-service returned status %d looking up student role", resp.StatusCode)
	}
	var role struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
		return uuid.Nil, fmt.Errorf("decode student role: %w", err)
	}
	return uuid.Parse(role.ID)
}

func (s *AuthService) fetchRolePermissions(roleID *uuid.UUID) []string {
	if roleID == nil || *roleID == uuid.Nil {
		return nil
	}

	url := fmt.Sprintf("%s/api/v1/roles/%s/permissions", s.cfg.UserServiceURL, roleID.String())
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var perms []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&perms); err != nil {
		return nil
	}

	names := make([]string, len(perms))
	for i, p := range perms {
		names[i] = p.Name
	}
	return names
}
