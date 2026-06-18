package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/Archiit19/School-management/pkg/pagination"
	"github.com/Archiit19/School-management/student-service/internal/config"
	"github.com/Archiit19/School-management/student-service/internal/model"
	"github.com/Archiit19/School-management/student-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentService struct {
	repo         *repository.StudentRepository
	cfg          *config.Config
	userInternal *httpclient.Client
	outboundHTTP *http.Client
}

func NewStudentService(
	repo *repository.StudentRepository,
	cfg *config.Config,
	userInternal *httpclient.Client,
	outboundHTTP *http.Client,
) *StudentService {
	return &StudentService{
		repo:         repo,
		cfg:          cfg,
		userInternal: userInternal,
		outboundHTTP: outboundHTTP,
	}
}

func (s *StudentService) CreateStudent(
	req model.CreateStudentRequest,
	schoolID uuid.UUID,
	authHeader string,
) (*model.CreateStudentResponse, error) {
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

	classInfo, sectionInfo, err := s.getClassSectionInfo(authHeader, classID, sectionUUID)
	if err != nil {
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

	wantLogin, err := validateLoginFields(req)
	if err != nil {
		return nil, err
	}

	admissionYear := time.Now().Year()
	studentCode, err := s.generateStudentCode(schoolID, classInfo.Name, sectionInfo, admissionYear)
	if err != nil {
		return nil, fmt.Errorf("failed to generate student code: %w", err)
	}

	student := &model.Student{
		SchoolID:      schoolID,
		StudentCode:   studentCode,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		ParentName:    strings.TrimSpace(req.ParentName),
		ContactNumber: strings.TrimSpace(req.ContactNumber),
		ParentUserID:  parentUUID,
		ClassID:       classID,
		SectionID:     sectionUUID,
		AdmissionYear: admissionYear,
		IsActive:      true,
	}

	if err := s.repo.CreateStudent(student); err != nil {
		return nil, fmt.Errorf("failed to create student: %w", err)
	}

	resp := &model.CreateStudentResponse{Student: *student}

	if wantLogin {
		if err := s.provisionStudentLogin(student, req.LoginEmail, req.LoginPassword); err != nil {
			_ = s.repo.DeleteStudent(student.ID)
			return nil, fmt.Errorf("admission rolled back: %w", err)
		}
		resp.LoginCreated = true
		resp.LoginEmail = req.LoginEmail
	}

	return resp, nil
}

func validateLoginFields(req model.CreateStudentRequest) (bool, error) {
	hasEmail := strings.TrimSpace(req.LoginEmail) != ""
	hasPwd := strings.TrimSpace(req.LoginPassword) != ""
	if hasEmail != hasPwd {
		return false, errors.New("login_email and login_password must be provided together")
	}
	return hasEmail && hasPwd, nil
}

// provisionStudentLogin asks auth-service to create a pupil auth user linked to this student.
func (s *StudentService) provisionStudentLogin(student *model.Student, email, password string) error {
	if strings.TrimSpace(s.cfg.InternalServiceToken) == "" {
		return errors.New("login provisioning is not configured (set INTERNAL_SERVICE_TOKEN)")
	}

	body, err := json.Marshal(map[string]string{
		"school_id":  student.SchoolID.String(),
		"student_id": student.ID.String(),
		"name":       strings.TrimSpace(student.FirstName + " " + student.LastName),
		"email":      email,
		"password":   password,
	})
	if err != nil {
		return fmt.Errorf("encode login request: %w", err)
	}

	url := fmt.Sprintf("%s/internal/users/from-student", s.cfg.UserServiceURL)
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build login request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.userInternal.Do(httpReq)
	if err != nil {
		return fmt.Errorf("user-service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return nil
	}
	respBody, _ := io.ReadAll(resp.Body)
	var errResp struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
		return fmt.Errorf("user-service rejected login (status %d): %s", resp.StatusCode, errResp.Error)
	}
	return fmt.Errorf("user-service rejected login (status %d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
}

func (s *StudentService) GetStudents(
	schoolID uuid.UUID,
	query model.StudentListQuery,
) (*model.StudentListResponse, error) {
	params := pagination.Params{Page: query.Page, Limit: query.Limit}
	pagination.Normalize(&params, pagination.Options{})
	query.Page = params.Page
	query.Limit = params.Limit

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

func (s *StudentService) GetStudentByID(studentID uuid.UUID) (*model.Student, error) {
	student, err := s.repo.GetStudentByID(studentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("student not found")
		}
		return nil, fmt.Errorf("failed to fetch student: %w", err)
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
	if req.ParentName != nil {
		student.ParentName = strings.TrimSpace(*req.ParentName)
	}
	if req.ContactNumber != nil {
		student.ContactNumber = strings.TrimSpace(*req.ContactNumber)
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
	url := fmt.Sprintf("%s/users/%s", s.cfg.UserServiceURL, parentID.String())
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", authHeader)

	resp, err := s.outboundHTTP.Do(req)
	if err != nil {
		return errors.New("failed to validate parent user with user-service")
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
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"class"`
	Sections []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"sections"`
}

type classInfo struct {
	ID   string
	Name string
}

type sectionInfo struct {
	ID   string
	Name string
}

func (s *StudentService) getClassSectionInfo(
	authHeader string,
	classID uuid.UUID,
	sectionID *uuid.UUID,
) (*classInfo, *sectionInfo, error) {
	url := fmt.Sprintf("%s/classes", s.cfg.AcademicServiceURL)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", authHeader)

	resp, err := s.outboundHTTP.Do(req)
	if err != nil {
		return nil, nil, errors.New("failed to validate class/section with academic-service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("academic-service validation failed: %s", strings.TrimSpace(string(body)))
	}

	var classes []academicClassResponse
	if err := json.NewDecoder(resp.Body).Decode(&classes); err != nil {
		return nil, nil, errors.New("failed to decode academic-service response")
	}

	var foundClass *classInfo
	var foundSection *sectionInfo

	for _, class := range classes {
		if class.Class.ID != classID.String() {
			continue
		}
		foundClass = &classInfo{ID: class.Class.ID, Name: class.Class.Name}
		if sectionID == nil {
			break
		}
		for _, section := range class.Sections {
			if section.ID == sectionID.String() {
				foundSection = &sectionInfo{ID: section.ID, Name: section.Name}
				break
			}
		}
		break
	}

	if foundClass == nil {
		return nil, nil, errors.New("class not found")
	}
	if sectionID != nil && foundSection == nil {
		return nil, nil, errors.New("section does not belong to the selected class")
	}

	return foundClass, foundSection, nil
}

func (s *StudentService) validateClassSection(
	authHeader string,
	classID uuid.UUID,
	sectionID *uuid.UUID,
) error {
	_, _, err := s.getClassSectionInfo(authHeader, classID, sectionID)
	return err
}

func (s *StudentService) generateStudentCode(
	schoolID uuid.UUID,
	className string,
	section *sectionInfo,
	admissionYear int,
) (string, error) {
	classNum := extractClassNumber(className)
	sectionLetter := extractSectionLetter(section)

	codePrefix := fmt.Sprintf("%04d%s%s", admissionYear, classNum, sectionLetter)
	maxUsed, err := s.repo.MaxEnrollmentForCodePrefix(codePrefix)
	if err != nil {
		return "", err
	}
	enrollmentNum := maxUsed + 1

	return fmt.Sprintf("%s%02d", codePrefix, enrollmentNum), nil
}

func extractSectionLetter(section *sectionInfo) string {
	if section == nil {
		return "X"
	}
	name := strings.ToUpper(strings.TrimSpace(section.Name))
	if name == "" {
		return "X"
	}
	if len(name) == 1 {
		return name
	}
	parts := strings.Fields(name)
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if len(last) == 1 {
			return last
		}
	}
	return string([]rune(name)[0])
}

func extractClassNumber(className string) string {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(className)
	if match == "" {
		return "00"
	}
	num, err := strconv.Atoi(match)
	if err != nil {
		return "00"
	}
	return fmt.Sprintf("%02d", num)
}
