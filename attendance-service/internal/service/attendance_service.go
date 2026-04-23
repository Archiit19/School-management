package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/avaneeshravat/school-management/attendance-service/internal/apierrors"
	"github.com/avaneeshravat/school-management/attendance-service/internal/config"
	"github.com/avaneeshravat/school-management/attendance-service/internal/model"
	"github.com/avaneeshravat/school-management/attendance-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendanceService struct {
	repo       *repository.AttendanceRepository
	cfg        *config.Config
	httpClient *http.Client
}

func NewAttendanceService(repo *repository.AttendanceRepository, cfg *config.Config, httpClient *http.Client) *AttendanceService {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 8 * time.Second}
	}
	return &AttendanceService{repo: repo, cfg: cfg, httpClient: httpClient}
}

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "23505")
}

type authUserInternal struct {
	ID       string `json:"id"`
	SchoolID string `json:"school_id"`
	IsActive bool   `json:"is_active"`
	RoleName string `json:"role_name"`
}

func (s *AttendanceService) validateAuthUserInSchool(userID, schoolID uuid.UUID) error {
	if strings.TrimSpace(s.cfg.InternalServiceToken) == "" {
		return apierrors.ServiceUnavailable("user validation is not configured (set INTERNAL_SERVICE_TOKEN and AUTH_SERVICE_URL)")
	}
	base := strings.TrimRight(s.cfg.AuthServiceURL, "/")
	url := fmt.Sprintf("%s/internal/users/%s", base, userID.String())
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to build auth request: %w", err)
	}
	req.Header.Set("X-Internal-Token", s.cfg.InternalServiceToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return apierrors.ServiceUnavailable("auth-service unreachable for user validation")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return apierrors.BadRequest("user not found")
	}
	if resp.StatusCode != http.StatusOK {
		return apierrors.ServiceUnavailable("auth-service user validation failed")
	}

	var u authUserInternal
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return apierrors.ServiceUnavailable("invalid response from auth-service")
	}

	sid, err := uuid.Parse(u.SchoolID)
	if err != nil || sid != schoolID {
		return apierrors.Forbidden("user does not belong to this school")
	}
	if !u.IsActive {
		return apierrors.BadRequest("user account is inactive")
	}
	return nil
}

func (s *AttendanceService) CreateAttendance(
	req model.CreateAttendanceRequest,
	schoolID, teacherUserID uuid.UUID,
	roleName string,
) (*model.Attendance, error) {
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
		if isDuplicateKey(err) {
			return nil, apierrors.Conflict("attendance already marked for this student and date")
		}
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
			return nil, apierrors.NotFound("attendance not found")
		}
		return nil, fmt.Errorf("failed to fetch attendance: %w", err)
	}

	if roleName != "super_admin" && record.TeacherUserID != requestingUserID {
		return nil, apierrors.Forbidden("you can edit only your own attendance entries")
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

func (s *AttendanceService) BulkCreateAttendance(
	req model.BulkCreateAttendanceRequest,
	schoolID, teacherUserID uuid.UUID,
	roleName string,
) (*model.BulkAttendanceResponse, error) {
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

	resp := &model.BulkAttendanceResponse{}

	for _, entry := range req.Entries {
		studentID, err := uuid.Parse(entry.StudentID)
		if err != nil {
			resp.Skipped++
			continue
		}

		status := normalizeStatus(entry.Status)
		if !isValidStatus(status) {
			resp.Skipped++
			continue
		}

		_, dupErr := s.repo.GetAttendanceByComposite(schoolID, studentID, classID, sectionID, subjectID, attendanceDate)
		if dupErr == nil {
			resp.Skipped++
			continue
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
			Remarks:       entry.Remarks,
		}
		if err := s.repo.CreateAttendance(record); err != nil {
			if isDuplicateKey(err) {
				resp.Skipped++
				continue
			}
			resp.Skipped++
			continue
		}

		resp.Created++
		resp.Records = append(resp.Records, *record)
	}

	return resp, nil
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

func hasPerm(perms []string, name string) bool {
	for _, p := range perms {
		if p == name {
			return true
		}
	}
	return false
}

func (s *AttendanceService) CreateTeacherAttendance(
	req model.CreateTeacherAttendanceRequest,
	schoolID, currentUserID uuid.UUID,
	roleName string,
	perms []string,
) (*model.TeacherAttendance, error) {
	var targetTeacherID uuid.UUID
	if strings.TrimSpace(req.TeacherUserID) == "" {
		targetTeacherID = currentUserID
	} else {
		parsed, err := uuid.Parse(req.TeacherUserID)
		if err != nil {
			return nil, errors.New("invalid teacher_user_id")
		}
		targetTeacherID = parsed
	}

	if err := s.assertCanMarkTeacherAttendance(roleName, perms, currentUserID, targetTeacherID); err != nil {
		return nil, err
	}

	if err := s.validateAuthUserInSchool(targetTeacherID, schoolID); err != nil {
		return nil, err
	}

	attendanceDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	status := normalizeStatus(req.Status)
	if !isValidStatus(status) {
		return nil, errors.New("invalid status, allowed: present, absent, late, excused")
	}

	_, err = s.repo.GetTeacherAttendanceBySchoolTeacherDate(schoolID, targetTeacherID, attendanceDate)
	if err == nil {
		return nil, errors.New("attendance already marked for this teacher and date")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to validate teacher attendance uniqueness: %w", err)
	}

	record := &model.TeacherAttendance{
		SchoolID:         schoolID,
		TeacherUserID:    targetTeacherID,
		RecordedByUserID: currentUserID,
		Date:             attendanceDate,
		Status:           status,
		Remarks:          req.Remarks,
	}
	if err := s.repo.CreateTeacherAttendance(record); err != nil {
		if isDuplicateKey(err) {
			return nil, apierrors.Conflict("attendance already marked for this teacher and date")
		}
		return nil, fmt.Errorf("failed to create teacher attendance: %w", err)
	}

	return record, nil
}

func (s *AttendanceService) assertCanMarkTeacherAttendance(
	roleName string,
	perms []string,
	currentUserID, targetTeacherID uuid.UUID,
) error {
	if roleName == "super_admin" {
		return nil
	}
	hasMarkAll := hasPerm(perms, "mark_teacher_attendance")
	hasMarkOwn := hasPerm(perms, "mark_own_teacher_attendance")
	if targetTeacherID == currentUserID {
		if hasMarkAll || hasMarkOwn {
			return nil
		}
		return apierrors.Forbidden("cannot mark your own attendance with current permissions")
	}
	if hasMarkAll {
		return nil
	}
	return apierrors.Forbidden("can only mark your own attendance")
}

func (s *AttendanceService) BulkCreateTeacherAttendance(
	req model.BulkCreateTeacherAttendanceRequest,
	schoolID, currentUserID uuid.UUID,
	roleName string,
	perms []string,
) (*model.BulkTeacherAttendanceResponse, error) {
	if roleName != "super_admin" && !hasPerm(perms, "mark_teacher_attendance") {
		return nil, apierrors.Forbidden("bulk teacher attendance requires mark_teacher_attendance")
	}

	attendanceDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	resp := &model.BulkTeacherAttendanceResponse{}

	for _, entry := range req.Entries {
		teacherID, err := uuid.Parse(entry.TeacherUserID)
		if err != nil {
			resp.Skipped++
			continue
		}

		if err := s.validateAuthUserInSchool(teacherID, schoolID); err != nil {
			resp.Skipped++
			continue
		}

		status := normalizeStatus(entry.Status)
		if !isValidStatus(status) {
			resp.Skipped++
			continue
		}

		_, dupErr := s.repo.GetTeacherAttendanceBySchoolTeacherDate(schoolID, teacherID, attendanceDate)
		if dupErr == nil {
			resp.Skipped++
			continue
		}
		if dupErr != nil && !errors.Is(dupErr, gorm.ErrRecordNotFound) {
			resp.Skipped++
			continue
		}

		record := &model.TeacherAttendance{
			SchoolID:         schoolID,
			TeacherUserID:    teacherID,
			RecordedByUserID: currentUserID,
			Date:             attendanceDate,
			Status:           status,
			Remarks:          entry.Remarks,
		}
		if err := s.repo.CreateTeacherAttendance(record); err != nil {
			if isDuplicateKey(err) {
				resp.Skipped++
				continue
			}
			resp.Skipped++
			continue
		}

		resp.Created++
		resp.Records = append(resp.Records, *record)
	}

	return resp, nil
}

func (s *AttendanceService) GetTeacherAttendance(
	schoolID, currentUserID uuid.UUID,
	roleName string,
	perms []string,
	query model.TeacherAttendanceQuery,
) (*model.TeacherAttendanceListResponse, error) {
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

	if roleName != "super_admin" && !hasPerm(perms, "view_teacher_attendance") && !hasPerm(perms, "mark_teacher_attendance") {
		if !hasPerm(perms, "mark_own_teacher_attendance") {
			return nil, apierrors.Forbidden("cannot view teacher attendance")
		}
		query.TeacherUserID = currentUserID.String()
	}

	records, total, err := s.repo.GetTeacherAttendance(schoolID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch teacher attendance: %w", err)
	}

	return &model.TeacherAttendanceListResponse{
		Attendance: records,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
	}, nil
}

func (s *AttendanceService) UpdateTeacherAttendance(
	id uuid.UUID,
	req model.UpdateTeacherAttendanceRequest,
	schoolID, requestingUserID uuid.UUID,
	roleName string,
) (*model.TeacherAttendance, error) {
	record, err := s.repo.GetTeacherAttendanceByIDAndSchool(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierrors.NotFound("teacher attendance not found")
		}
		return nil, fmt.Errorf("failed to fetch teacher attendance: %w", err)
	}

	if roleName != "super_admin" && record.RecordedByUserID != requestingUserID {
		return nil, apierrors.Forbidden("you can edit only your own teacher attendance entries")
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

	if err := s.repo.UpdateTeacherAttendance(record); err != nil {
		return nil, fmt.Errorf("failed to update teacher attendance: %w", err)
	}

	return record, nil
}
