package service

import (
	"context"
	"errors"
	"fmt"

	log "github.com/Archiit19/School-management/pkg/logger"
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
		log.Error("create school: email check failed", log.Err(err), log.AddField("email", req.Email))
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
		log.Error("create school: database insert failed", log.Err(err), log.AddField("email", req.Email))
		return nil, fmt.Errorf("failed to create school: %w", err)
	}
	log.Info("school created", log.AddField("school_id", school.ID), log.AddField("name", school.Name))
	return school, nil
}

func (s *SchoolService) CreateSchoolForUser(ctx context.Context, userID uuid.UUID, req model.CreateSchoolRequest) (*model.School, error) {
	school, err := s.CreateSchool(req)
	if err != nil {
		return nil, err
	}

	if err := s.bootstrap.BootstrapSchool(ctx, school.ID); err != nil {
		return nil, fmt.Errorf("failed to bootstrap school roles: %w", err)
	}

	roleID, err := s.bootstrap.FetchRoleID(ctx, school.ID, "super_admin")
	if err != nil {
		log.Error("create school for user: resolve super_admin role failed", log.Err(err), log.AddField("school_id", school.ID), log.AddField("user_id", userID))
		return nil, fmt.Errorf("failed to resolve super_admin role: %w", err)
	}

	if err := s.repo.CreateMembership(&model.UserSchool{SchoolID: school.ID, UserID: userID}); err != nil {
		log.Error("create school for user: membership insert failed", log.Err(err), log.AddField("school_id", school.ID), log.AddField("user_id", userID))
		return nil, fmt.Errorf("failed to link user to school: %w", err)
	}

	if err := s.bootstrap.AssignUserRole(ctx, userID, school.ID, roleID); err != nil {
		return nil, fmt.Errorf("failed to assign super_admin role: %w", err)
	}

	log.Info("school created for user", log.AddField("school_id", school.ID), log.AddField("user_id", userID))
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
		log.Error("add member: membership lookup failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("user_id", userID))
		return nil, err
	}

	m := &model.UserSchool{SchoolID: schoolID, UserID: userID}
	if err := s.repo.CreateMembership(m); err != nil {
		log.Error("add member: database insert failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("user_id", userID))
		return nil, fmt.Errorf("failed to add member: %w", err)
	}
	log.Info("school member added", log.AddField("school_id", schoolID), log.AddField("user_id", userID))
	return m, nil
}

func (s *SchoolService) RemoveMember(schoolID, userID uuid.UUID) error {
	if err := s.repo.DeleteMembership(schoolID, userID); err != nil {
		log.Error("remove member: database delete failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("user_id", userID))
		return err
	}
	log.Info("school member removed", log.AddField("school_id", schoolID), log.AddField("user_id", userID))
	return nil
}

func (s *SchoolService) GetMembership(schoolID, userID uuid.UUID) (*model.UserSchool, error) {
	m, err := s.repo.GetMembership(schoolID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("membership not found")
		}
		log.Error("get membership: database query failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("user_id", userID))
		return nil, err
	}
	return m, nil
}

func (s *SchoolService) ListMembershipsForUser(userID uuid.UUID) ([]model.UserSchoolMember, error) {
	rows, err := s.repo.GetMembershipsForUser(userID)
	if err != nil {
		log.Error("list memberships for user: database query failed", log.Err(err), log.AddField("user_id", userID))
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
		log.Error("list members for school: database query failed", log.Err(err), log.AddField("school_id", schoolID))
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
		log.Error("get school: database query failed", log.Err(err), log.AddField("school_id", id))
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
		log.Error("get school by email: database query failed", log.Err(err), log.AddField("email", email))
		return nil, fmt.Errorf("failed to fetch school: %w", err)
	}
	return school, nil
}

func (s *SchoolService) ListSchools(query model.SchoolListQuery) (*model.SchoolListResponse, error) {
	schools, total, err := s.repo.List(query)
	if err != nil {
		log.Error("list schools: database query failed", log.Err(err))
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
	schools, err := s.repo.ListSchoolsForUser(userID)
	if err != nil {
		log.Error("list schools for user: database query failed", log.Err(err), log.AddField("user_id", userID))
		return nil, err
	}
	return schools, nil
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
		log.Error("update school: database fetch failed", log.Err(err), log.AddField("school_id", id))
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
				log.Error("update school: email check failed", log.Err(err), log.AddField("school_id", id))
				return nil, fmt.Errorf("failed to check email: %w", err)
			}
		}
		school.Email = *req.Email
	}
	if req.IsActive != nil {
		school.IsActive = *req.IsActive
	}

	if err := s.repo.Update(school); err != nil {
		log.Error("update school: database update failed", log.Err(err), log.AddField("school_id", id))
		return nil, fmt.Errorf("failed to update school: %w", err)
	}
	log.Info("school updated", log.AddField("school_id", school.ID))
	return school, nil
}

func (s *SchoolService) DeleteSchool(id uuid.UUID) error {
	school, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("school not found")
		}
		log.Error("delete school: database fetch failed", log.Err(err), log.AddField("school_id", id))
		return fmt.Errorf("failed to fetch school: %w", err)
	}
	school.IsActive = false
	if err := s.repo.Update(school); err != nil {
		log.Error("delete school: database update failed", log.Err(err), log.AddField("school_id", id))
		return err
	}
	log.Info("school deactivated", log.AddField("school_id", id))
	return nil
}
