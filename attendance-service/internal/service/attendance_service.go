package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/avaneeshravat/school-management/attendance-service/internal/model"
	"github.com/avaneeshravat/school-management/attendance-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendanceService struct {
	repo *repository.AttendanceRepository
}

func NewAttendanceService(repo *repository.AttendanceRepository) *AttendanceService {
	return &AttendanceService{repo: repo}
}

func (s *AttendanceService) CreateAttendance(
	req model.CreateAttendanceRequest,
	schoolID, teacherUserID uuid.UUID,
	roleName string,
) (*model.Attendance, error) {
	if roleName != "teacher" && roleName != "super_admin" {
		return nil, errors.New("only teacher or super_admin can mark attendance")
	}

	studentID, err := uuid.Parse(req.StudentID)
	if err != nil {
		return nil, errors.New("invalid student_id")
	}
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

	attendanceDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	status := normalizeStatus(req.Status)
	if !isValidStatus(status) {
		return nil, errors.New("invalid status, allowed: present, absent, late, excused")
	}

	_, err = s.repo.GetAttendanceByComposite(schoolID, studentID, classID, sectionID, subjectID, attendanceDate)
	if err == nil {
		return nil, errors.New("attendance already marked for this student and date")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to validate attendance uniqueness: %w", err)
	}

	record := &model.Attendance{
		SchoolID:      schoolID,
		TeacherUserID: teacherUserID,
		StudentID:     studentID,
		ClassID:       classID,
		SectionID:     sectionID,
		SubjectID:     subjectID,
		Date:          attendanceDate,
		Status:        status,
		Remarks:       req.Remarks,
	}
	if err := s.repo.CreateAttendance(record); err != nil {
		return nil, fmt.Errorf("failed to create attendance: %w", err)
	}

	return record, nil
}

func (s *AttendanceService) GetAttendance(
	schoolID uuid.UUID,
	query model.AttendanceQuery,
) (*model.AttendanceListResponse, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	if strings.TrimSpace(query.Date) != "" {
		if _, err := time.Parse("2006-01-02", query.Date); err != nil {
			return nil, errors.New("invalid date format, use YYYY-MM-DD")
		}
	}

	records, total, err := s.repo.GetAttendance(schoolID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attendance: %w", err)
	}

	return &model.AttendanceListResponse{
		Attendance: records,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
	}, nil
}

func (s *AttendanceService) UpdateAttendance(
	id uuid.UUID,
	req model.UpdateAttendanceRequest,
	schoolID, requestingUserID uuid.UUID,
	roleName string,
) (*model.Attendance, error) {
	record, err := s.repo.GetAttendanceByIDAndSchool(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("attendance not found")
		}
		return nil, fmt.Errorf("failed to fetch attendance: %w", err)
	}

	if roleName != "super_admin" && record.TeacherUserID != requestingUserID {
		return nil, errors.New("you can edit only your own attendance entries")
	}

	if req.Status != nil {
		status := normalizeStatus(*req.Status)
		if !isValidStatus(status) {
			return nil, errors.New("invalid status, allowed: present, absent, late, excused")
		}
		record.Status = status
	}

	if req.Remarks != nil {
		record.Remarks = *req.Remarks
	}

	if err := s.repo.UpdateAttendance(record); err != nil {
		return nil, fmt.Errorf("failed to update attendance: %w", err)
	}

	return record, nil
}

func normalizeStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

func isValidStatus(status string) bool {
	switch status {
	case "present", "absent", "late", "excused":
		return true
	default:
		return false
	}
}
