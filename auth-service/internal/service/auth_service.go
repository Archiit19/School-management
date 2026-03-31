package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/avaneeshravat/school-management/auth-service/internal/config"
	"github.com/avaneeshravat/school-management/auth-service/internal/model"
	"github.com/avaneeshravat/school-management/auth-service/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	repo *repository.AuthRepository
	cfg  *config.Config
}

func NewAuthService(repo *repository.AuthRepository, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

// ─── Register School ────────────────────────────────────────────────

func (s *AuthService) RegisterSchool(req model.RegisterSchoolRequest) (*model.RegisterSchoolResponse, error) {
	// 1. Check if school email already exists
	_, err := s.repo.GetSchoolByEmail(req.SchoolEmail)
	if err == nil {
		return nil, errors.New("school with this email already exists")
	}

	// 2. Check if admin email already exists
	_, err = s.repo.GetUserByEmail(req.AdminEmail)
	if err == nil {
		return nil, errors.New("user with this email already exists")
	}

	// 3. Create school
	school := &model.School{
		Name:    req.SchoolName,
		Address: req.SchoolAddress,
		Phone:   req.SchoolPhone,
		Email:   req.SchoolEmail,
	}
	if err := s.repo.CreateSchool(school); err != nil {
		return nil, fmt.Errorf("failed to create school: %w", err)
	}

	// 4. Call User Service to create "super_admin" role for this school
	roleID, err := s.createDefaultRole(school.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create default role via user-service: %w", err)
	}

	// 5. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 6. Create admin user
	admin := &model.User{
		SchoolID: school.ID,
		Name:     req.AdminName,
		Email:    req.AdminEmail,
		Password: string(hashedPassword),
		RoleID:   roleID,
		IsActive: true,
	}
	if err := s.repo.CreateUser(admin); err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	// 7. Generate JWT (with role_name for RBAC middleware)
	token, err := s.generateToken(admin, "super_admin")
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	admin.RoleName = "super_admin"

	return &model.RegisterSchoolResponse{
		School: *school,
		Admin:  *admin,
		Token:  token,
	}, nil
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

	// Fetch role name from user service
	roleName := s.fetchRoleName(user.RoleID)
	user.RoleName = roleName

	token, err := s.generateToken(user, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &model.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

// ─── Get Current User (Me) ─────────────────────────────────────────

func (s *AuthService) GetMe(userID uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Enrich with role name
	roleName := s.fetchRoleName(user.RoleID)
	user.RoleName = roleName

	return user, nil
}

// ─── JWT Helpers ────────────────────────────────────────────────────

func (s *AuthService) generateToken(user *model.User, roleName string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   user.ID.String(),
		"school_id": user.SchoolID.String(),
		"role_id":   user.RoleID.String(),
		"role_name": roleName,
		"email":     user.Email,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// ─── Inter-Service: Create Default Role ─────────────────────────────

// createDefaultRole calls the User Service to create a "super_admin" role
// for the newly registered school and returns the role ID.
func (s *AuthService) createDefaultRole(schoolID uuid.UUID) (uuid.UUID, error) {
	payload := map[string]string{
		"name":        "super_admin",
		"description": "Super Administrator with full access",
		"school_id":   schoolID.String(),
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/api/v1/roles/internal", s.cfg.UserServiceURL)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return uuid.Nil, fmt.Errorf("user-service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return uuid.Nil, fmt.Errorf("user-service returned status %d", resp.StatusCode)
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return uuid.Nil, fmt.Errorf("failed to decode user-service response: %w", err)
	}

	return uuid.Parse(result.ID)
}

// fetchRoleName calls the User Service to get the role name by ID.
func (s *AuthService) fetchRoleName(roleID uuid.UUID) string {
	if roleID == uuid.Nil {
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
