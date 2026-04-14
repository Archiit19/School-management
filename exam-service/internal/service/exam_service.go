package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/avaneeshravat/school-management/exam-service/internal/model"
	"github.com/avaneeshravat/school-management/exam-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExamService struct {
	repo *repository.ExamRepository
}

func NewExamService(repo *repository.ExamRepository) *ExamService {
	return &ExamService{repo: repo}
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
