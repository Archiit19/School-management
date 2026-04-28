package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/avaneeshravat/school-management/student-service/internal/config"
	"github.com/avaneeshravat/school-management/student-service/internal/model"
	"github.com/avaneeshravat/school-management/student-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentService struct {
	repo       *repository.StudentRepository
	cfg        *config.Config
	httpClient *http.Client
}

func NewStudentService(
	repo *repository.StudentRepository,
	cfg *config.Config,
	httpClient *http.Client,
) *StudentService {
	return &StudentService{
		repo:       repo,
		cfg:        cfg,
		httpClient: httpClient,
	}
}

func (s *StudentService) CreateStudent(
	req model.CreateStudentRequest,
	schoolID uuid.UUID,
	authHeader string,
) (*model.Student, error) {
	classID, err := uuid.Parse(req.ClassID)
	if err != nil {
		return nil, errors.New("invalid class_id")
	}

	var sectionUUID *uuid.UUID
	if req.SectionID != "" {
		parsed, err := uuid.Parse(req.SectionID)
		if err != nil {
			return nil, errors.New("invalid section_id")
		}
		sectionUUID = &parsed
	}

	if err := s.validateClassSection(authHeader, classID, sectionUUID); err != nil {
		return nil, err
	}

	var parentUUID *uuid.UUID
	if req.ParentUserID != "" {
		parsedParentID, err := uuid.Parse(req.ParentUserID)
		if err != nil {
			return nil, errors.New("invalid parent_user_id")
		}

		if err := s.validateParent(authHeader, parsedParentID, schoolID); err != nil {
			return nil, err
		}
		parentUUID = &parsedParentID
	}

	student := &model.Student{
		SchoolID:     schoolID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		ParentUserID: parentUUID,
		ClassID:      classID,
		SectionID:    sectionUUID,
		IsActive:     true,
	}

	if err := s.repo.CreateStudent(student); err != nil {
		return nil, fmt.Errorf("failed to create student: %w", err)
	}

	return student, nil
}

func (s *StudentService) GetStudents(
	schoolID uuid.UUID,
	query model.StudentListQuery,
) (*model.StudentListResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	students, total, err := s.repo.GetStudentsBySchoolID(schoolID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch students: %w", err)
	}

	return &model.StudentListResponse{
		Students: students,
		Total:    total,
		Page:     query.Page,
		Limit:    query.Limit,
	}, nil
}

func (s *StudentService) GetStudentMe(schoolID, studentID uuid.UUID) (*model.Student, error) {
	student, err := s.repo.GetStudentByIDAndSchoolID(studentID, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("student not found")
		}
		return nil, fmt.Errorf("failed to fetch student: %w", err)
	}
	if !student.IsActive {
		return nil, errors.New("student account inactive")
	}
	return student, nil
}

func (s *StudentService) UpdateStudent(
	id uuid.UUID,
	req model.UpdateStudentRequest,
	schoolID uuid.UUID,
	authHeader string,
) (*model.Student, error) {
	student, err := s.repo.GetStudentByIDAndSchoolID(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("student not found")
		}
		return nil, fmt.Errorf("failed to fetch student: %w", err)
	}

	if req.FirstName != nil {
		student.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		student.LastName = *req.LastName
	}
	if req.IsActive != nil {
		student.IsActive = *req.IsActive
	}

	if req.ParentUserID != nil {
		if strings.TrimSpace(*req.ParentUserID) == "" {
			student.ParentUserID = nil
		} else {
			parentID, err := uuid.Parse(*req.ParentUserID)
			if err != nil {
				return nil, errors.New("invalid parent_user_id")
			}
			if err := s.validateParent(authHeader, parentID, schoolID); err != nil {
				return nil, err
			}
			student.ParentUserID = &parentID
		}
	}

	newClassID := student.ClassID
	if req.ClassID != nil {
		parsedClassID, err := uuid.Parse(*req.ClassID)
		if err != nil {
			return nil, errors.New("invalid class_id")
		}
		newClassID = parsedClassID
	}

	newSectionID := student.SectionID
	if req.SectionID != nil {
		if strings.TrimSpace(*req.SectionID) == "" {
			newSectionID = nil
		} else {
			parsedSectionID, err := uuid.Parse(*req.SectionID)
			if err != nil {
				return nil, errors.New("invalid section_id")
			}
			newSectionID = &parsedSectionID
		}
	}

	if req.ClassID != nil || req.SectionID != nil {
		if err := s.validateClassSection(authHeader, newClassID, newSectionID); err != nil {
			return nil, err
		}
		student.ClassID = newClassID
		student.SectionID = newSectionID
	}

	if err := s.repo.UpdateStudent(student); err != nil {
		return nil, fmt.Errorf("failed to update student: %w", err)
	}

	return student, nil
}

type authUserResponse struct {
	ID       string `json:"id"`
	SchoolID string `json:"school_id"`
	RoleName string `json:"role_name"`
}

func (s *StudentService) validateParent(authHeader string, parentID, schoolID uuid.UUID) error {
	url := fmt.Sprintf("%s/users/%s", s.cfg.AuthServiceURL, parentID.String())
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", authHeader)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return errors.New("failed to validate parent user with auth-service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("parent user not found or inaccessible")
	}

	var user authUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return errors.New("failed to decode parent user response")
	}

	parsedSchoolID, err := uuid.Parse(user.SchoolID)
	if err != nil || parsedSchoolID != schoolID {
		return errors.New("parent must belong to the same school")
	}

	if user.RoleName != "parent" {
		return errors.New("linked user must have parent role")
	}

	return nil
}

type academicClassResponse struct {
	Class struct {
		ID string `json:"id"`
	} `json:"class"`
	Sections []struct {
		ID string `json:"id"`
	} `json:"sections"`
}

func (s *StudentService) validateClassSection(
	authHeader string,
	classID uuid.UUID,
	sectionID *uuid.UUID,
) error {
	url := fmt.Sprintf("%s/classes", s.cfg.AcademicServiceURL)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", authHeader)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return errors.New("failed to validate class/section with academic-service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("academic-service validation failed: %s", strings.TrimSpace(string(body)))
	}

	var classes []academicClassResponse
	if err := json.NewDecoder(resp.Body).Decode(&classes); err != nil {
		return errors.New("failed to decode academic-service response")
	}

	classFound := false
	sectionFound := sectionID == nil
	for _, class := range classes {
		if class.Class.ID != classID.String() {
			continue
		}
		classFound = true
		if sectionID == nil {
			break
		}
		for _, section := range class.Sections {
			if section.ID == sectionID.String() {
				sectionFound = true
				break
			}
		}
		break
	}

	if !classFound {
		return errors.New("class not found")
	}
	if !sectionFound {
		return errors.New("section does not belong to the selected class")
	}

	return nil
}
