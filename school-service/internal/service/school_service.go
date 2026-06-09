package service

import (
	"errors"
	"fmt"

	"github.com/Archiit19/School-management/school-service/internal/model"
	"github.com/Archiit19/School-management/school-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SchoolService struct {
	repo      *repository.SchoolRepository
	bootstrap *authBootstrapClient
}

func NewSchoolService(repo *repository.SchoolRepository, authServiceURL, internalToken string) *SchoolService {
	return &SchoolService{
		repo:      repo,
		bootstrap: newAuthBootstrapClient(authServiceURL, internalToken),
	}
}

func (s *SchoolService) CreateSchool(req model.CreateSchoolRequest) (*model.School, error) {
	_, err := s.repo.GetByEmail(req.Email)
	if err == nil {
		return nil, errors.New("school with this email already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check school email: %w", err)
	}

	school := &model.School{
		Name:     req.Name,
		Address:  req.Address,
		Phone:    req.Phone,
		Email:    req.Email,
		IsActive: true,
	}
	if err := s.repo.Create(school); err != nil {
		return nil, fmt.Errorf("failed to create school: %w", err)
	}
	return school, nil
}

func (s *SchoolService) CreateSchoolForUser(userID uuid.UUID, req model.CreateSchoolRequest) (*model.School, error) {
	school, err := s.CreateSchool(req)
	if err != nil {
		return nil, err
	}

	if err := s.bootstrap.BootstrapSchool(school.ID); err != nil {
		return nil, fmt.Errorf("failed to bootstrap school roles: %w", err)
	}

	roleID, err := s.bootstrap.FetchRoleID(school.ID, "super_admin")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve super_admin role: %w", err)
	}

	if err := s.repo.CreateMembership(&model.UserSchool{SchoolID: school.ID, UserID: userID}); err != nil {
		return nil, fmt.Errorf("failed to link user to school: %w", err)
	}

	if err := s.bootstrap.AssignUserRole(userID, school.ID, roleID); err != nil {
		return nil, fmt.Errorf("failed to assign super_admin role: %w", err)
	}

	return school, nil
}

func (s *SchoolService) AddMember(schoolID, userID uuid.UUID) (*model.UserSchool, error) {
	if _, err := s.repo.GetByID(schoolID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("school not found")
		}
		return nil, err
	}
	if _, err := s.repo.GetMembership(schoolID, userID); err == nil {
		return nil, errors.New("user is already a member of this school")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	m := &model.UserSchool{SchoolID: schoolID, UserID: userID}
	if err := s.repo.CreateMembership(m); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}
	return m, nil
}

func (s *SchoolService) RemoveMember(schoolID, userID uuid.UUID) error {
	return s.repo.DeleteMembership(schoolID, userID)
}

func (s *SchoolService) GetMembership(schoolID, userID uuid.UUID) (*model.UserSchool, error) {
	m, err := s.repo.GetMembership(schoolID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("membership not found")
		}
		return nil, err
	}
	return m, nil
}

func (s *SchoolService) ListMembershipsForUser(userID uuid.UUID) ([]model.UserSchoolMember, error) {
	rows, err := s.repo.GetMembershipsForUser(userID)
	if err != nil {
		return nil, err
	}
	out := make([]model.UserSchoolMember, len(rows))
	for i, r := range rows {
		out[i] = model.UserSchoolMember{UserID: r.UserID, SchoolID: r.SchoolID}
	}
	return out, nil
}

func (s *SchoolService) ListMembersForSchool(schoolID uuid.UUID) ([]model.UserSchoolMember, error) {
	rows, err := s.repo.GetMembersForSchool(schoolID)
	if err != nil {
		return nil, err
	}
	out := make([]model.UserSchoolMember, len(rows))
	for i, r := range rows {
		out[i] = model.UserSchoolMember{UserID: r.UserID, SchoolID: r.SchoolID}
	}
	return out, nil
}

func (s *SchoolService) ListUserIDsForSchool(schoolID uuid.UUID) ([]uuid.UUID, error) {
	return s.repo.ListUserIDsForSchool(schoolID)
}

func (s *SchoolService) GetSchool(id uuid.UUID) (*model.School, error) {
	school, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("school not found")
		}
		return nil, fmt.Errorf("failed to fetch school: %w", err)
	}
	return school, nil
}

func (s *SchoolService) GetSchoolByEmail(email string) (*model.School, error) {
	school, err := s.repo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("school not found")
		}
		return nil, fmt.Errorf("failed to fetch school: %w", err)
	}
	return school, nil
}

func (s *SchoolService) ListSchools(query model.SchoolListQuery) (*model.SchoolListResponse, error) {
	schools, total, err := s.repo.List(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list schools: %w", err)
	}
	return &model.SchoolListResponse{
		Schools: schools,
		Total:   total,
		Page:    query.Page,
		Limit:   query.Limit,
	}, nil
}

func (s *SchoolService) ListSchoolsForUser(userID uuid.UUID) ([]model.School, error) {
	return s.repo.ListSchoolsForUser(userID)
}

func (s *SchoolService) IsUserMember(schoolID, userID uuid.UUID) (bool, error) {
	return s.repo.IsUserMember(schoolID, userID)
}

func (s *SchoolService) UpdateSchool(id uuid.UUID, req model.UpdateSchoolRequest) (*model.School, error) {
	school, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("school not found")
		}
		return nil, fmt.Errorf("failed to fetch school: %w", err)
	}

	if req.Name != nil {
		school.Name = *req.Name
	}
	if req.Address != nil {
		school.Address = *req.Address
	}
	if req.Phone != nil {
		school.Phone = *req.Phone
	}
	if req.Email != nil {
		if *req.Email != school.Email {
			if _, err := s.repo.GetByEmail(*req.Email); err == nil {
				return nil, errors.New("school with this email already exists")
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("failed to check email: %w", err)
			}
		}
		school.Email = *req.Email
	}
	if req.IsActive != nil {
		school.IsActive = *req.IsActive
	}

	if err := s.repo.Update(school); err != nil {
		return nil, fmt.Errorf("failed to update school: %w", err)
	}
	return school, nil
}

func (s *SchoolService) DeleteSchool(id uuid.UUID) error {
	school, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("school not found")
		}
		return fmt.Errorf("failed to fetch school: %w", err)
	}
	school.IsActive = false
	return s.repo.Update(school)
}
