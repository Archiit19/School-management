package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Archiit19/School-management/exam-service/internal/config"
	"github.com/Archiit19/School-management/exam-service/internal/model"
	"github.com/Archiit19/School-management/exam-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExamService struct {
	repo         *repository.ExamRepository
	cfg          *config.Config
	outboundHTTP *http.Client
}

func NewExamService(
	repo *repository.ExamRepository,
	cfg *config.Config,
	outboundHTTP *http.Client,
) *ExamService {
	return &ExamService{repo: repo, cfg: cfg, outboundHTTP: outboundHTTP}
}

func (s *ExamService) CreateExam(
	req model.CreateExamRequest,
	schoolID, createdBy uuid.UUID,
	roleName string,
) (*model.Exam, error) {
	classID, err := uuid.Parse(req.ClassID)
	if err != nil {
		return nil, errors.New("invalid class_id")
	}

	var sectionID *uuid.UUID
	if strings.TrimSpace(req.SectionID) != "" {
		parsed, err := uuid.Parse(req.SectionID)
		if err != nil {
			return nil, errors.New("invalid section_id")
		}
		sectionID = &parsed
	}

	var subjectID *uuid.UUID
	if strings.TrimSpace(req.SubjectID) != "" {
		parsed, err := uuid.Parse(req.SubjectID)
		if err != nil {
			return nil, errors.New("invalid subject_id")
		}
		subjectID = &parsed
	}

	examDate, err := time.Parse("2006-01-02", req.ExamDate)
	if err != nil {
		return nil, errors.New("invalid exam_date format, use YYYY-MM-DD")
	}

	exam := &model.Exam{
		SchoolID:    schoolID,
		ClassID:     classID,
		SectionID:   sectionID,
		SubjectID:   subjectID,
		Title:       req.Title,
		ExamDate:    examDate,
		TotalMarks:  req.TotalMarks,
		IsPublished: false,
		CreatedBy:   createdBy,
	}
	if err := s.repo.CreateExam(exam); err != nil {
		return nil, fmt.Errorf("failed to create exam: %w", err)
	}
	return exam, nil
}

func (s *ExamService) EnterMarks(
	req model.EnterMarksRequest,
	schoolID, createdBy uuid.UUID,
	roleName string,
) (*model.Mark, error) {
	examID, err := uuid.Parse(req.ExamID)
	if err != nil {
		return nil, errors.New("invalid exam_id")
	}
	studentID, err := uuid.Parse(req.StudentID)
	if err != nil {
		return nil, errors.New("invalid student_id")
	}

	exam, err := s.repo.GetExamByIDAndSchool(examID, schoolID)
	if err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.IsPublished {
		return nil, errors.New("cannot enter marks after results are published")
	}
	if req.MarksObtained < 0 || req.MarksObtained > exam.TotalMarks {
		return nil, errors.New("marks_obtained must be between 0 and total_marks")
	}

	existing, err := s.repo.GetMarkByExamAndStudent(examID, studentID)
	if err == nil {
		existing.MarksObtained = req.MarksObtained
		existing.Remarks = req.Remarks
		if err := s.repo.UpdateMark(existing); err != nil {
			return nil, fmt.Errorf("failed to update marks: %w", err)
		}
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to validate existing mark: %w", err)
	}

	mark := &model.Mark{
		SchoolID:      schoolID,
		ExamID:        examID,
		StudentID:     studentID,
		MarksObtained: req.MarksObtained,
		Remarks:       req.Remarks,
		CreatedBy:     createdBy,
	}
	if err := s.repo.CreateMark(mark); err != nil {
		return nil, fmt.Errorf("failed to create marks: %w", err)
	}
	return mark, nil
}

func (s *ExamService) PublishResults(
	req model.PublishResultRequest,
	schoolID uuid.UUID,
	roleName string,
) (*model.Exam, error) {
	examID, err := uuid.Parse(req.ExamID)
	if err != nil {
		return nil, errors.New("invalid exam_id")
	}

	exam, err := s.repo.GetExamByIDAndSchool(examID, schoolID)
	if err != nil {
		return nil, errors.New("exam not found")
	}
	exam.IsPublished = true
	if err := s.repo.UpdateExam(exam); err != nil {
		return nil, fmt.Errorf("failed to publish results: %w", err)
	}
	return exam, nil
}

// GetExams lists exams for a school. When upcoming=true, only exams from today
// onwards are returned.
func (s *ExamService) GetExams(schoolID uuid.UUID, query model.ExamQuery) ([]model.Exam, error) {
	exams, err := s.repo.GetExams(schoolID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exams: %w", err)
	}
	return exams, nil
}

// GetMyExams returns exams scheduled for the pupil's own class only.
// It resolves the pupil's class_id and section_id by calling user-service /users/me
// with the pupil's JWT, so spoofing another student is impossible.
func (s *ExamService) GetMyExams(ctx context.Context, 
	schoolID, studentID uuid.UUID,
	authHeader string,
	upcoming bool,
) ([]model.Exam, error) {
	if strings.TrimSpace(authHeader) == "" {
		return nil, errors.New("missing authorization header")
	}
	url := fmt.Sprintf(
		"%s/enrollments/me?student_id=%s",
		strings.TrimRight(s.cfg.AcademicServiceURL, "/"),
		studentID.String(),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := s.outboundHTTP.Do(req)
	if err != nil {
		return nil, errors.New("failed to resolve pupil enrollment from academic-service")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("academic-service /enrollments/me returned status %d", resp.StatusCode)
	}

	var enrollment struct {
		UserID    uuid.UUID  `json:"user_id"`
		ClassID   uuid.UUID  `json:"class_id"`
		SectionID *uuid.UUID `json:"section_id,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&enrollment); err != nil {
		return nil, errors.New("failed to decode enrollment response")
	}
	if enrollment.UserID != studentID {
		return nil, errors.New("pupil profile mismatch")
	}

	var pupilSectionID *uuid.UUID = enrollment.SectionID

	query := model.ExamQuery{
		ClassID:  enrollment.ClassID.String(),
		Upcoming: upcoming,
	}
	exams, err := s.repo.GetExams(schoolID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exams: %w", err)
	}

	filtered := make([]model.Exam, 0, len(exams))
	for _, e := range exams {
		if e.SectionID == nil || pupilSectionID == nil || *e.SectionID == *pupilSectionID {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil
}

func (s *ExamService) GetResults(
	schoolID uuid.UUID,
	query model.ResultQuery,
	roleName string,
) ([]model.ResultItem, error) {
	includeUnpublished := roleName == "teacher" || roleName == "super_admin"
	results, err := s.repo.GetResults(schoolID, query, includeUnpublished)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch results: %w", err)
	}

	for i := range results {
		if results[i].TotalMarks <= 0 {
			results[i].Percentage = 0
			results[i].Grade = "N/A"
			continue
		}
		results[i].Percentage = (results[i].MarksObtained / results[i].TotalMarks) * 100
		results[i].Grade = gradeFromPercentage(results[i].Percentage)
	}
	return results, nil
}

func gradeFromPercentage(p float64) string {
	switch {
	case p >= 90:
		return "A+"
	case p >= 80:
		return "A"
	case p >= 70:
		return "B"
	case p >= 60:
		return "C"
	case p >= 50:
		return "D"
	default:
		return "F"
	}
}
