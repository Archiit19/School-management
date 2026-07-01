package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/Archiit19/School-management/pkg/logger"
	"github.com/Archiit19/School-management/academic-service/internal/config"
	"github.com/Archiit19/School-management/academic-service/internal/model"
	"github.com/Archiit19/School-management/academic-service/internal/repository"
	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AcademicService struct {
	repo         *repository.AcademicRepository
	cfg          *config.Config
	userInternal *httpclient.Client
	outboundHTTP *http.Client
}

func NewAcademicService(
	repo *repository.AcademicRepository,
	cfg *config.Config,
	userInternal *httpclient.Client,
	outboundHTTP *http.Client,
) *AcademicService {
	return &AcademicService{
		repo:         repo,
		cfg:          cfg,
		userInternal: userInternal,
		outboundHTTP: outboundHTTP,
	}
}

func (s *AcademicService) CreateClass(req model.CreateClassRequest, schoolID uuid.UUID) (*model.Class, error) {
	_, err := s.repo.GetClassByName(schoolID, req.Name)
	if err == nil {
		return nil, errors.New("class with this name already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("create class: uniqueness check failed", log.Err(err), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to validate class uniqueness: %w", err)
	}

	class := &model.Class{
		SchoolID:    schoolID,
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.repo.CreateClass(class); err != nil {
		log.Error("create class: database insert failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("name", req.Name))
		return nil, fmt.Errorf("failed to create class: %w", err)
	}
	log.Info("class created", log.AddField("class_id", class.ID), log.AddField("school_id", schoolID), log.AddField("name", class.Name))
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
		log.Error("create section: uniqueness check failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("class_id", classID))
		return nil, fmt.Errorf("failed to validate section uniqueness: %w", err)
	}

	section := &model.Section{
		SchoolID: schoolID,
		ClassID:  classID,
		Name:     req.Name,
	}
	if err := s.repo.CreateSection(section); err != nil {
		log.Error("create section: database insert failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("class_id", classID))
		return nil, fmt.Errorf("failed to create section: %w", err)
	}
	log.Info("section created", log.AddField("section_id", section.ID), log.AddField("school_id", schoolID), log.AddField("class_id", classID))
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
		log.Error("create subject: uniqueness check failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("class_id", classID))
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
		log.Error("create subject: database insert failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("class_id", classID))
		return nil, fmt.Errorf("failed to create subject: %w", err)
	}
	log.Info("subject created", log.AddField("subject_id", subject.ID), log.AddField("school_id", schoolID), log.AddField("class_id", classID))
	return subject, nil
}

func (s *AcademicService) GetClasses(schoolID uuid.UUID) ([]model.ClassWithChildren, error) {
	classes, err := s.repo.GetClassesBySchoolID(schoolID)
	if err != nil {
		log.Error("list classes: database query failed", log.Err(err), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to fetch classes: %w", err)
	}

	resp := make([]model.ClassWithChildren, 0, len(classes))
	for _, class := range classes {
		sections, err := s.repo.GetSectionsByClassID(class.ID)
		if err != nil {
			log.Error("list classes: sections query failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("class_id", class.ID))
			return nil, fmt.Errorf("failed to fetch sections for class %s: %w", class.ID, err)
		}
		subjects, err := s.repo.GetSubjectsByClassID(class.ID)
		if err != nil {
			log.Error("list classes: subjects query failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("class_id", class.ID))
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

func (s *AcademicService) CreateTeacherAssignment(ctx context.Context, 
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

	if err := s.validateTeacher(ctx, authHeader, teacherUserID, schoolID); err != nil {
		return nil, err
	}

	_, err = s.repo.GetTeacherAssignmentByClassSubject(schoolID, classID, subjectID)
	if err == nil {
		return nil, errors.New("this subject already has a teacher assigned for this class")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("create teacher assignment: subject assignment check failed", log.Err(err), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to validate subject assignment: %w", err)
	}

	_, err = s.repo.GetTeacherAssignmentByComposite(schoolID, teacherUserID, classID, subjectID)
	if err == nil {
		return nil, errors.New("teacher assignment already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("create teacher assignment: uniqueness check failed", log.Err(err), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to validate assignment uniqueness: %w", err)
	}

	assignment := &model.TeacherAssignment{
		SchoolID:      schoolID,
		TeacherUserID: teacherUserID,
		ClassID:       classID,
		SubjectID:     subjectID,
	}
	if err := s.repo.CreateTeacherAssignment(assignment); err != nil {
		log.Error("create teacher assignment: database insert failed", log.Err(err), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to create teacher assignment: %w", err)
	}
	log.Info("teacher assignment created",
		log.AddField("teacher_assignment_id", assignment.ID),
		log.AddField("school_id", schoolID),
		log.AddField("teacher_user_id", teacherUserID),
	)

	return assignment, nil
}

func (s *AcademicService) GetTeacherAssignments(
	schoolID uuid.UUID,
	query model.TeacherAssignmentQuery,
) ([]model.TeacherAssignment, error) {
	assignments, err := s.repo.GetTeacherAssignments(schoolID, query)
	if err != nil {
		log.Error("list teacher assignments: database query failed", log.Err(err), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to fetch teacher assignments: %w", err)
	}
	return assignments, nil
}

func (s *AcademicService) UpdateTeacherAssignment(ctx context.Context, 
	id uuid.UUID,
	req model.UpdateTeacherAssignmentRequest,
	schoolID uuid.UUID,
	authHeader string,
) (*model.TeacherAssignment, error) {
	assignment, err := s.repo.GetTeacherAssignmentByIDAndSchool(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("teacher assignment not found")
		}
		log.Error("update teacher assignment: database fetch failed", log.Err(err), log.AddField("teacher_assignment_id", id), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to fetch teacher assignment: %w", err)
	}

	teacherUserID := assignment.TeacherUserID
	classID := assignment.ClassID
	subjectID := assignment.SubjectID

	if req.TeacherUserID != nil && strings.TrimSpace(*req.TeacherUserID) != "" {
		parsed, err := uuid.Parse(*req.TeacherUserID)
		if err != nil {
			return nil, errors.New("invalid teacher_user_id")
		}
		teacherUserID = parsed
	}
	if req.ClassID != nil && strings.TrimSpace(*req.ClassID) != "" {
		parsed, err := uuid.Parse(*req.ClassID)
		if err != nil {
			return nil, errors.New("invalid class_id")
		}
		classID = parsed
	}
	if req.SubjectID != nil && strings.TrimSpace(*req.SubjectID) != "" {
		parsed, err := uuid.Parse(*req.SubjectID)
		if err != nil {
			return nil, errors.New("invalid subject_id")
		}
		subjectID = parsed
	}

	if req.TeacherUserID == nil && req.ClassID == nil && req.SubjectID == nil {
		return nil, errors.New("at least one field must be provided to update")
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

	if err := s.validateTeacher(ctx, authHeader, teacherUserID, schoolID); err != nil {
		return nil, err
	}

	existingBySubject, err := s.repo.GetTeacherAssignmentByClassSubject(schoolID, classID, subjectID)
	if err == nil && existingBySubject.ID != id {
		return nil, errors.New("this subject already has a teacher assigned for this class")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to validate subject assignment: %w", err)
	}

	existingComposite, err := s.repo.GetTeacherAssignmentByComposite(schoolID, teacherUserID, classID, subjectID)
	if err == nil && existingComposite.ID != id {
		return nil, errors.New("teacher assignment already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to validate assignment uniqueness: %w", err)
	}

	assignment.TeacherUserID = teacherUserID
	assignment.ClassID = classID
	assignment.SubjectID = subjectID

	if err := s.repo.UpdateTeacherAssignment(assignment); err != nil {
		log.Error("update teacher assignment: database update failed", log.Err(err), log.AddField("teacher_assignment_id", id), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to update teacher assignment: %w", err)
	}
	log.Info("teacher assignment updated", log.AddField("teacher_assignment_id", assignment.ID), log.AddField("school_id", schoolID))

	return assignment, nil
}

func (s *AcademicService) DeleteTeacherAssignment(id, schoolID uuid.UUID) error {
	_, err := s.repo.GetTeacherAssignmentByIDAndSchool(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("teacher assignment not found")
		}
		log.Error("delete teacher assignment: database fetch failed", log.Err(err), log.AddField("teacher_assignment_id", id), log.AddField("school_id", schoolID))
		return fmt.Errorf("failed to fetch teacher assignment: %w", err)
	}
	if err := s.repo.DeleteTeacherAssignment(id, schoolID); err != nil {
		log.Error("delete teacher assignment: database delete failed", log.Err(err), log.AddField("teacher_assignment_id", id), log.AddField("school_id", schoolID))
		return fmt.Errorf("failed to delete teacher assignment: %w", err)
	}
	log.Info("teacher assignment deleted", log.AddField("teacher_assignment_id", id), log.AddField("school_id", schoolID))
	return nil
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
		log.Error("create assignment: database insert failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("title", req.Title))
		return nil, fmt.Errorf("failed to create assignment: %w", err)
	}
	log.Info("assignment created",
		log.AddField("assignment_id", assignment.ID),
		log.AddField("school_id", schoolID),
		log.AddField("title", assignment.Title),
	)

	return assignment, nil
}

func (s *AcademicService) GetAssignments(
	schoolID uuid.UUID,
	query model.AssignmentQuery,
) ([]model.Assignment, error) {
	assignments, err := s.repo.GetAssignments(schoolID, query)
	if err != nil {
		log.Error("list assignments: database query failed", log.Err(err), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to fetch assignments: %w", err)
	}
	return assignments, nil
}

// GetMyAssignments resolves the pupil's class via user-service then lists assignments for that class only.
func (s *AcademicService) GetMyAssignments(
	schoolID, studentID uuid.UUID,
	authHeader string,
) ([]model.Assignment, error) {
	st, err := s.resolvePupilProfile(studentID, schoolID)
	if err != nil {
		return nil, err
	}
	return s.GetAssignments(schoolID, model.AssignmentQuery{ClassID: st.ClassID.String()})
}

// GetMyAcademicProfile returns the pupil's class, section, subjects, and the
// teachers assigned to that class. Class/section are resolved by calling
// user-service /users/me with the pupil's JWT; teacher names are resolved
// via auth-service's internal /internal/users/:id endpoint.
func (s *AcademicService) GetMyAcademicProfile(ctx context.Context, 
	schoolID, studentID uuid.UUID,
	authHeader string,
) (*model.MyAcademicProfile, error) {
	st, err := s.resolvePupilProfile(studentID, schoolID)
	if err != nil {
		return nil, err
	}

	profile := &model.MyAcademicProfile{
		Subjects: []model.Subject{},
		Teachers: []model.ClassTeacher{},
	}

	// 2. Class
	if class, err := s.repo.GetClassByIDAndSchool(st.ClassID, schoolID); err == nil {
		profile.Class = class
	}

	// 3. Section (optional)
	if st.SectionID != nil {
		if sec, err := s.repo.GetSectionByIDAndSchool(*st.SectionID, schoolID); err == nil {
			profile.Section = sec
		}
	}

	// 4. Subjects for the pupil's class, scoped to their section (+ class-wide subjects).
	if subjects, err := s.repo.GetSubjectsByClassID(st.ClassID); err == nil {
		profile.Subjects = filterSubjectsForStudent(subjects, st.SectionID)
	}

	// 5. Teacher assignments for the class, then enrich names via auth-service.
	tas, err := s.repo.GetTeacherAssignments(schoolID, model.TeacherAssignmentQuery{
		ClassID: st.ClassID.String(),
	})
	if err != nil {
		return profile, nil // partial response is still useful
	}

	subjectName := make(map[uuid.UUID]string, len(profile.Subjects))
	visibleSubjectIDs := make(map[uuid.UUID]struct{}, len(profile.Subjects))
	for _, sub := range profile.Subjects {
		subjectName[sub.ID] = sub.Name
		visibleSubjectIDs[sub.ID] = struct{}{}
	}

	for _, ta := range tas {
		if _, ok := visibleSubjectIDs[ta.SubjectID]; !ok {
			continue
		}
		ct := model.ClassTeacher{
			TeacherUserID: ta.TeacherUserID,
			SubjectID:     ta.SubjectID,
			SubjectName:   subjectName[ta.SubjectID],
		}
		if name, email, err := s.resolveUser(ctx, ta.TeacherUserID); err == nil {
			ct.TeacherName = name
			ct.TeacherEmail = email
		}
		profile.Teachers = append(profile.Teachers, ct)
	}
	return profile, nil
}

// filterSubjectsForStudent returns subjects the pupil should see: class-wide subjects
// (no section) plus subjects tied to the pupil's own section. Without this filter,
// every section's copy of "Hindi", "Math", etc. would appear and look like duplicates.
func filterSubjectsForStudent(subjects []model.Subject, sectionID *uuid.UUID) []model.Subject {
	filtered := make([]model.Subject, 0, len(subjects))
	for _, s := range subjects {
		if s.SectionID == nil {
			filtered = append(filtered, s)
			continue
		}
		if sectionID != nil && *s.SectionID == *sectionID {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// resolveUser fetches a user's name and email via auth-service's internal endpoint.
// Returns empty strings (no error) when the call is not configured or fails — the
// caller treats missing names as "Unknown teacher" but still surfaces the subject.
func (s *AcademicService) resolveUser(ctx context.Context, userID uuid.UUID) (string, string, error) {
	if strings.TrimSpace(s.cfg.InternalServiceToken) == "" {
		return "", "", errors.New("internal service token not configured")
	}
	url := fmt.Sprintf("/internal/users/%s", userID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.userInternal.URL(url), nil)
	if err != nil {
		return "", "", err
	}
	resp, err := s.userInternal.DoContext(ctx, req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("user-service /internal/users returned %d", resp.StatusCode)
	}
	var u struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return "", "", err
	}
	return u.Name, u.Email, nil
}

func (s *AcademicService) GetMySubmissions(schoolID, studentID uuid.UUID) ([]model.Submission, error) {
	subs, err := s.repo.GetSubmissionsForStudent(schoolID, studentID)
	if err != nil {
		log.Error("get my submissions: database query failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("student_id", studentID))
		return nil, fmt.Errorf("failed to fetch submissions: %w", err)
	}
	return subs, nil
}

// CreateOwnSubmission lets a pupil submit; student_id and submitted_by are forced from JWT.
func (s *AcademicService) CreateOwnSubmission(
	req model.CreateMySubmissionRequest,
	schoolID, studentID uuid.UUID,
) (*model.Submission, error) {
	assignmentID, err := uuid.Parse(req.AssignmentID)
	if err != nil {
		return nil, errors.New("invalid assignment_id")
	}

	if _, err := s.repo.GetAssignmentByIDAndSchool(assignmentID, schoolID); err != nil {
		return nil, errors.New("assignment not found")
	}

	if _, err := s.repo.GetSubmissionByComposite(schoolID, assignmentID, studentID); err == nil {
		return nil, errors.New("submission already exists for this assignment")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("create own submission: uniqueness check failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("assignment_id", assignmentID))
		return nil, fmt.Errorf("failed to validate submission uniqueness: %w", err)
	}

	submission := &model.Submission{
		SchoolID:     schoolID,
		AssignmentID: assignmentID,
		StudentID:    studentID,
		SubmittedBy:  studentID,
		Content:      req.Content,
		MaterialURL:  req.MaterialURL,
	}
	if err := s.repo.CreateSubmission(submission); err != nil {
		log.Error("create own submission: database insert failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("assignment_id", assignmentID))
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}
	log.Info("submission created",
		log.AddField("submission_id", submission.ID),
		log.AddField("school_id", schoolID),
		log.AddField("assignment_id", assignmentID),
		log.AddField("student_id", studentID),
	)
	return submission, nil
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
		log.Error("create submission: uniqueness check failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("assignment_id", assignmentID))
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
		log.Error("create submission: database insert failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("assignment_id", assignmentID))
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}
	log.Info("submission created",
		log.AddField("submission_id", submission.ID),
		log.AddField("school_id", schoolID),
		log.AddField("assignment_id", assignmentID),
		log.AddField("student_id", studentID),
	)

	return submission, nil
}

func (s *AcademicService) canManageAssignmentSubmissions(
	assignment *model.Assignment,
	schoolID, userID uuid.UUID,
	roleName string,
) error {
	if roleName == "super_admin" {
		return nil
	}
	if roleName != "teacher" {
		return errors.New("only teacher or super_admin can review submissions")
	}
	if assignment.TeacherUserID == userID {
		return nil
	}
	if _, err := s.repo.GetTeacherAssignmentByComposite(schoolID, userID, assignment.ClassID, assignment.SubjectID); err == nil {
		return nil
	}
	return errors.New("you are not allowed to access submissions for this assignment")
}

// resolveStudent returns display info for a pupil. submission.student_id stores the
// pupil's user-service user ID (see CreateOwnSubmission), not student-service admission IDs.
func (s *AcademicService) resolveStudent(ctx context.Context, studentUserID uuid.UUID) (string, string, error) {
	name, _, err := s.resolveUser(ctx, studentUserID)
	if err != nil {
		return "", "", err
	}
	code, _ := s.resolveStudentCode(ctx, studentUserID)
	return name, code, nil
}

func (s *AcademicService) resolveStudentCode(ctx context.Context, userID uuid.UUID) (string, error) {
	if strings.TrimSpace(s.cfg.InternalServiceToken) == "" {
		return "", errors.New("internal service token not configured")
	}
	url := fmt.Sprintf("/internal/users/%s/profile", userID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.userInternal.URL(url), nil)
	if err != nil {
		return "", err
	}
	resp, err := s.userInternal.DoContext(ctx, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("user-service /internal/users/profile returned %d", resp.StatusCode)
	}
	var profile map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return "", err
	}
	if v, ok := profile["student_code"]; ok {
		if code := strings.TrimSpace(fmt.Sprint(v)); code != "" {
			return code, nil
		}
	}
	return "", nil
}

func (s *AcademicService) enrichSubmissionView(ctx context.Context, sub model.Submission) model.SubmissionView {
	view := model.SubmissionView{Submission: sub}
	if name, code, err := s.resolveStudent(ctx, sub.StudentID); err == nil {
		view.StudentName = name
		view.StudentCode = code
	}
	if name, _, err := s.resolveUser(ctx, sub.SubmittedBy); err == nil {
		view.SubmitterName = name
	}
	return view
}

func (s *AcademicService) GetSubmissionsForAssignment(ctx context.Context, 
	assignmentID, schoolID, userID uuid.UUID,
	roleName string,
) ([]model.SubmissionView, error) {
	assignment, err := s.repo.GetAssignmentByIDAndSchool(assignmentID, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("assignment not found")
		}
		log.Error("list assignment submissions: assignment fetch failed", log.Err(err), log.AddField("assignment_id", assignmentID), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to fetch assignment: %w", err)
	}
	if err := s.canManageAssignmentSubmissions(assignment, schoolID, userID, roleName); err != nil {
		return nil, err
	}

	subs, err := s.repo.GetSubmissionsForAssignment(schoolID, assignmentID)
	if err != nil {
		log.Error("list assignment submissions: database query failed", log.Err(err), log.AddField("assignment_id", assignmentID), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to fetch submissions: %w", err)
	}
	out := make([]model.SubmissionView, len(subs))
	for i, sub := range subs {
		out[i] = s.enrichSubmissionView(ctx, sub)
	}
	return out, nil
}

func (s *AcademicService) ReviewSubmission(
	submissionID, schoolID, userID uuid.UUID,
	roleName string,
	req model.UpdateSubmissionRequest,
) (*model.Submission, error) {
	submission, err := s.repo.GetSubmissionByIDAndSchool(submissionID, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("submission not found")
		}
		log.Error("review submission: database fetch failed", log.Err(err), log.AddField("submission_id", submissionID), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to fetch submission: %w", err)
	}

	assignment, err := s.repo.GetAssignmentByIDAndSchool(submission.AssignmentID, schoolID)
	if err != nil {
		return nil, errors.New("assignment not found")
	}
	if err := s.canManageAssignmentSubmissions(assignment, schoolID, userID, roleName); err != nil {
		return nil, err
	}

	if req.TeacherFeedback != nil {
		submission.TeacherFeedback = strings.TrimSpace(*req.TeacherFeedback)
	}
	if req.Marks != nil {
		if *req.Marks < 0 || *req.Marks > 20 {
			return nil, errors.New("marks must be between 0 and 20")
		}
		marks := *req.Marks
		submission.Marks = &marks
	}
	now := time.Now()
	submission.ReviewedAt = &now

	if err := s.repo.UpdateSubmission(submission); err != nil {
		log.Error("review submission: database update failed", log.Err(err), log.AddField("submission_id", submissionID), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to save review: %w", err)
	}
	log.Info("submission reviewed",
		log.AddField("submission_id", submission.ID),
		log.AddField("school_id", schoolID),
		log.AddField("assignment_id", submission.AssignmentID),
	)
	return submission, nil
}

type authUserResponse struct {
	ID       string `json:"id"`
	SchoolID string `json:"school_id"`
	RoleName string `json:"role_name"`
}

func (s *AcademicService) validateTeacher(ctx context.Context, 
	authHeader string,
	teacherUserID, schoolID uuid.UUID,
) error {
	url := fmt.Sprintf("%s/users/%s", s.cfg.UserServiceURL, teacherUserID.String())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Authorization", authHeader)

	resp, err := s.outboundHTTP.Do(req)
	if err != nil {
		return errors.New("failed to validate teacher user with user-service")
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

type pupilProfile struct {
	ID        uuid.UUID
	ClassID   uuid.UUID
	SectionID *uuid.UUID
}

func (s *AcademicService) resolvePupilProfile(studentID, schoolID uuid.UUID) (*pupilProfile, error) {
	enrollment, err := s.repo.GetEnrollmentByUserAndSchool(studentID, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("pupil enrollment not found")
		}
		log.Error("resolve pupil profile: enrollment fetch failed", log.Err(err), log.AddField("student_id", studentID), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to load enrollment: %w", err)
	}
	return &pupilProfile{
		ID:        enrollment.UserID,
		ClassID:   enrollment.ClassID,
		SectionID: enrollment.SectionID,
	}, nil
}

func (s *AcademicService) ListEnrollments(schoolID uuid.UUID, query model.EnrollmentQuery) (*model.EnrollmentListResponse, error) {
	classID, err := uuid.Parse(query.ClassID)
	if err != nil {
		return nil, errors.New("invalid class_id")
	}
	var sectionID *uuid.UUID
	if query.SectionID != "" {
		parsed, err := uuid.Parse(query.SectionID)
		if err != nil {
			return nil, errors.New("invalid section_id")
		}
		sectionID = &parsed
	}
	if _, err := s.repo.GetClassByIDAndSchool(classID, schoolID); err != nil {
		return nil, errors.New("class not found")
	}
	if sectionID != nil {
		if _, err := s.repo.GetSectionByIDAndSchool(*sectionID, schoolID); err != nil {
			return nil, errors.New("section not found")
		}
	}
	rows, err := s.repo.ListEnrollmentsByClassSection(schoolID, classID, sectionID)
	if err != nil {
		log.Error("list enrollments: database query failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("class_id", classID))
		return nil, fmt.Errorf("failed to list enrollments: %w", err)
	}
	return &model.EnrollmentListResponse{Enrollments: rows, Total: int64(len(rows))}, nil
}

func (s *AcademicService) GetEnrollmentByUser(userID, schoolID uuid.UUID) (*model.StudentEnrollment, error) {
	enrollment, err := s.repo.GetEnrollmentByUserAndSchool(userID, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("enrollment not found")
		}
		log.Error("get enrollment: database query failed", log.Err(err), log.AddField("user_id", userID), log.AddField("school_id", schoolID))
		return nil, err
	}
	return enrollment, nil
}

func (s *AcademicService) UpsertEnrollment(req model.UpsertEnrollmentRequest) (*model.StudentEnrollment, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user_id")
	}
	schoolID, err := uuid.Parse(req.SchoolID)
	if err != nil {
		return nil, errors.New("invalid school_id")
	}
	classID, err := uuid.Parse(req.ClassID)
	if err != nil {
		return nil, errors.New("invalid class_id")
	}
	if _, err := s.repo.GetClassByIDAndSchool(classID, schoolID); err != nil {
		return nil, errors.New("class not found")
	}
	var sectionID *uuid.UUID
	if req.SectionID != "" {
		parsed, err := uuid.Parse(req.SectionID)
		if err != nil {
			return nil, errors.New("invalid section_id")
		}
		if _, err := s.repo.GetSectionByIDAndSchool(parsed, schoolID); err != nil {
			return nil, errors.New("section not found")
		}
		sectionID = &parsed
	}
	row := &model.StudentEnrollment{
		SchoolID:  schoolID,
		UserID:    userID,
		ClassID:   classID,
		SectionID: sectionID,
		IsActive:  true,
	}
	if err := s.repo.UpsertEnrollment(row); err != nil {
		log.Error("upsert enrollment: database save failed", log.Err(err), log.AddField("user_id", userID), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to save enrollment: %w", err)
	}
	enrollment, err := s.repo.GetEnrollmentByUserAndSchool(userID, schoolID)
	if err != nil {
		log.Error("upsert enrollment: reload failed", log.Err(err), log.AddField("user_id", userID), log.AddField("school_id", schoolID))
		return nil, err
	}
	log.Info("enrollment upserted",
		log.AddField("user_id", userID),
		log.AddField("school_id", schoolID),
		log.AddField("class_id", classID),
	)
	return enrollment, nil
}

func (s *AcademicService) DeleteEnrollment(userID, schoolID uuid.UUID) error {
	if err := s.repo.DeleteEnrollment(userID, schoolID); err != nil {
		log.Error("delete enrollment: database delete failed", log.Err(err), log.AddField("user_id", userID), log.AddField("school_id", schoolID))
		return err
	}
	log.Info("enrollment deleted", log.AddField("user_id", userID), log.AddField("school_id", schoolID))
	return nil
}
