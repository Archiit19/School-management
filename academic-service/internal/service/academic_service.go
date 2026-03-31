package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/avaneeshravat/school-management/academic-service/internal/config"
	"github.com/avaneeshravat/school-management/academic-service/internal/model"
	"github.com/avaneeshravat/school-management/academic-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AcademicService struct {
	repo       *repository.AcademicRepository
	cfg        *config.Config
	httpClient *http.Client
}

func NewAcademicService(
	repo *repository.AcademicRepository,
	cfg *config.Config,
	httpClient *http.Client,
) *AcademicService {
	return &AcademicService{
		repo:       repo,
		cfg:        cfg,
		httpClient: httpClient,
	}
}

func (s *AcademicService) CreateClass(req model.CreateClassRequest, schoolID uuid.UUID) (*model.Class, error) {
	_, err := s.repo.GetClassByName(schoolID, req.Name)
	if err == nil {
		return nil, errors.New("class with this name already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to validate class uniqueness: %w", err)
	}

	class := &model.Class{
		SchoolID:    schoolID,
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.repo.CreateClass(class); err != nil {
		return nil, fmt.Errorf("failed to create class: %w", err)
	}
	return class, nil
}

func (s *AcademicService) CreateSection(req model.CreateSectionRequest, schoolID uuid.UUID) (*model.Section, error) {
	classID, err := uuid.Parse(req.ClassID)
	if err != nil {
		return nil, errors.New("invalid class_id")
	}

	_, err = s.repo.GetClassByIDAndSchool(classID, schoolID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	_, err = s.repo.GetSectionByClassAndName(classID, req.Name)
	if err == nil {
		return nil, errors.New("section with this name already exists in this class")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to validate section uniqueness: %w", err)
	}

	section := &model.Section{
		SchoolID: schoolID,
		ClassID:  classID,
		Name:     req.Name,
	}
	if err := s.repo.CreateSection(section); err != nil {
		return nil, fmt.Errorf("failed to create section: %w", err)
	}
	return section, nil
}

func (s *AcademicService) CreateSubject(req model.CreateSubjectRequest, schoolID uuid.UUID) (*model.Subject, error) {
	classID, err := uuid.Parse(req.ClassID)
	if err != nil {
		return nil, errors.New("invalid class_id")
	}

	_, err = s.repo.GetClassByIDAndSchool(classID, schoolID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	var parsedSectionID *uuid.UUID
	if req.SectionID != "" {
		sectionID, err := uuid.Parse(req.SectionID)
		if err != nil {
			return nil, errors.New("invalid section_id")
		}

		section, err := s.repo.GetSectionByIDAndSchool(sectionID, schoolID)
		if err != nil {
			return nil, errors.New("section not found")
		}
		if section.ClassID != classID {
			return nil, errors.New("section does not belong to the given class")
		}
		parsedSectionID = &sectionID
	}

	_, err = s.repo.GetSubjectByNameAndScope(schoolID, classID, parsedSectionID, req.Name)
	if err == nil {
		return nil, errors.New("subject with this name already exists in this scope")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to validate subject uniqueness: %w", err)
	}

	subject := &model.Subject{
		SchoolID:  schoolID,
		ClassID:   classID,
		SectionID: parsedSectionID,
		Name:      req.Name,
		Code:      req.Code,
	}
	if err := s.repo.CreateSubject(subject); err != nil {
		return nil, fmt.Errorf("failed to create subject: %w", err)
	}
	return subject, nil
}

func (s *AcademicService) GetClasses(schoolID uuid.UUID) ([]model.ClassWithChildren, error) {
	classes, err := s.repo.GetClassesBySchoolID(schoolID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch classes: %w", err)
	}

	resp := make([]model.ClassWithChildren, 0, len(classes))
	for _, class := range classes {
		sections, err := s.repo.GetSectionsByClassID(class.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch sections for class %s: %w", class.ID, err)
		}
		subjects, err := s.repo.GetSubjectsByClassID(class.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch subjects for class %s: %w", class.ID, err)
		}

		resp = append(resp, model.ClassWithChildren{
			Class:    class,
			Sections: sections,
			Subjects: subjects,
		})
	}
	return resp, nil
}

func (s *AcademicService) CreateTeacherAssignment(
	req model.CreateTeacherAssignmentRequest,
	schoolID uuid.UUID,
	authHeader string,
) (*model.TeacherAssignment, error) {
	teacherUserID, err := uuid.Parse(req.TeacherUserID)
	if err != nil {
		return nil, errors.New("invalid teacher_user_id")
	}
	classID, err := uuid.Parse(req.ClassID)
	if err != nil {
		return nil, errors.New("invalid class_id")
	}
	subjectID, err := uuid.Parse(req.SubjectID)
	if err != nil {
		return nil, errors.New("invalid subject_id")
	}

	_, err = s.repo.GetClassByIDAndSchool(classID, schoolID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	subject, err := s.repo.GetSubjectByIDAndSchool(subjectID, schoolID)
	if err != nil {
		return nil, errors.New("subject not found")
	}
	if subject.ClassID != classID {
		return nil, errors.New("subject does not belong to the selected class")
	}

	if err := s.validateTeacher(authHeader, teacherUserID, schoolID); err != nil {
		return nil, err
	}

	_, err = s.repo.GetTeacherAssignmentByComposite(schoolID, teacherUserID, classID, subjectID)
	if err == nil {
		return nil, errors.New("teacher assignment already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to validate assignment uniqueness: %w", err)
	}

	assignment := &model.TeacherAssignment{
		SchoolID:      schoolID,
		TeacherUserID: teacherUserID,
		ClassID:       classID,
		SubjectID:     subjectID,
	}
	if err := s.repo.CreateTeacherAssignment(assignment); err != nil {
		return nil, fmt.Errorf("failed to create teacher assignment: %w", err)
	}

	return assignment, nil
}

func (s *AcademicService) GetTeacherAssignments(
	schoolID uuid.UUID,
	query model.TeacherAssignmentQuery,
) ([]model.TeacherAssignment, error) {
	assignments, err := s.repo.GetTeacherAssignments(schoolID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch teacher assignments: %w", err)
	}
	return assignments, nil
}

func (s *AcademicService) CreateAssignment(
	req model.CreateAssignmentRequest,
	schoolID, requestingUserID uuid.UUID,
	roleName string,
) (*model.Assignment, error) {
	if roleName != "teacher" && roleName != "super_admin" {
		return nil, errors.New("only teacher or super_admin can create assignments")
	}

	classID, err := uuid.Parse(req.ClassID)
	if err != nil {
		return nil, errors.New("invalid class_id")
	}
	subjectID, err := uuid.Parse(req.SubjectID)
	if err != nil {
		return nil, errors.New("invalid subject_id")
	}

	_, err = s.repo.GetClassByIDAndSchool(classID, schoolID)
	if err != nil {
		return nil, errors.New("class not found")
	}
	subject, err := s.repo.GetSubjectByIDAndSchool(subjectID, schoolID)
	if err != nil {
		return nil, errors.New("subject not found")
	}
	if subject.ClassID != classID {
		return nil, errors.New("subject does not belong to the selected class")
	}

	if roleName == "teacher" {
		_, err := s.repo.GetTeacherAssignmentByComposite(schoolID, requestingUserID, classID, subjectID)
		if err != nil {
			return nil, errors.New("teacher is not assigned to this class and subject")
		}
	}

	var dueDate *time.Time
	if strings.TrimSpace(req.DueDate) != "" {
		parsed, err := time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			return nil, errors.New("invalid due_date format, use YYYY-MM-DD")
		}
		dueDate = &parsed
	}

	assignment := &model.Assignment{
		SchoolID:      schoolID,
		TeacherUserID: requestingUserID,
		ClassID:       classID,
		SubjectID:     subjectID,
		Title:         req.Title,
		Description:   req.Description,
		MaterialURL:   req.MaterialURL,
		DueDate:       dueDate,
	}
	if err := s.repo.CreateAssignment(assignment); err != nil {
		return nil, fmt.Errorf("failed to create assignment: %w", err)
	}

	return assignment, nil
}

func (s *AcademicService) GetAssignments(
	schoolID uuid.UUID,
	query model.AssignmentQuery,
) ([]model.Assignment, error) {
	assignments, err := s.repo.GetAssignments(schoolID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch assignments: %w", err)
	}
	return assignments, nil
}

func (s *AcademicService) CreateSubmission(
	req model.CreateSubmissionRequest,
	schoolID, submittedBy uuid.UUID,
) (*model.Submission, error) {
	assignmentID, err := uuid.Parse(req.AssignmentID)
	if err != nil {
		return nil, errors.New("invalid assignment_id")
	}
	studentID, err := uuid.Parse(req.StudentID)
	if err != nil {
		return nil, errors.New("invalid student_id")
	}

	_, err = s.repo.GetAssignmentByIDAndSchool(assignmentID, schoolID)
	if err != nil {
		return nil, errors.New("assignment not found")
	}

	_, err = s.repo.GetSubmissionByComposite(schoolID, assignmentID, studentID)
	if err == nil {
		return nil, errors.New("submission already exists for this assignment and student")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to validate submission uniqueness: %w", err)
	}

	submission := &model.Submission{
		SchoolID:     schoolID,
		AssignmentID: assignmentID,
		StudentID:    studentID,
		SubmittedBy:  submittedBy,
		Content:      req.Content,
		MaterialURL:  req.MaterialURL,
	}
	if err := s.repo.CreateSubmission(submission); err != nil {
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}

	return submission, nil
}

type authUserResponse struct {
	ID       string `json:"id"`
	SchoolID string `json:"school_id"`
	RoleName string `json:"role_name"`
}

func (s *AcademicService) validateTeacher(
	authHeader string,
	teacherUserID, schoolID uuid.UUID,
) error {
	url := fmt.Sprintf("%s/users/%s", s.cfg.AuthServiceURL, teacherUserID.String())
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", authHeader)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return errors.New("failed to validate teacher user with auth-service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("teacher user not found or inaccessible")
	}

	var user authUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return errors.New("failed to decode teacher user response")
	}

	userSchoolID, err := uuid.Parse(user.SchoolID)
	if err != nil || userSchoolID != schoolID {
		return errors.New("teacher must belong to the same school")
	}

	if user.RoleName != "teacher" {
		return errors.New("linked user must have teacher role")
	}

	return nil
}
