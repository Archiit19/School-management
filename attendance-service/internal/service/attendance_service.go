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
	"github.com/Archiit19/School-management/pkg/apierrors"
	"github.com/Archiit19/School-management/pkg/pagination"
	"github.com/Archiit19/School-management/attendance-service/internal/config"
	"github.com/Archiit19/School-management/attendance-service/internal/model"
	"github.com/Archiit19/School-management/attendance-service/internal/repository"
	"github.com/Archiit19/School-management/pkg/httpclient"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendanceService struct {
	repo         *repository.AttendanceRepository
	cfg          *config.Config
	userInternal *httpclient.Client
	outboundHTTP *http.Client
}

func NewAttendanceService(
	repo *repository.AttendanceRepository,
	cfg *config.Config,
	userInternal *httpclient.Client,
	outboundHTTP *http.Client,
) *AttendanceService {
	return &AttendanceService{
		repo:         repo,
		cfg:          cfg,
		userInternal: userInternal,
		outboundHTTP: outboundHTTP,
	}
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

type teacherAssignmentRow struct {
	TeacherUserID string `json:"teacher_user_id"`
	ClassID       string `json:"class_id"`
	SubjectID     string `json:"subject_id"`
}

func (s *AttendanceService) fetchTeacherAssignments(ctx context.Context, 
	teacherUserID uuid.UUID,
	authHeader string,
) ([]teacherAssignmentRow, error) {
	base := strings.TrimRight(s.cfg.AcademicServiceURL, "/")
	url := fmt.Sprintf("%s/teacher-assignments?teacher_user_id=%s", base, teacherUserID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build academic request: %w", err)
	}
	if strings.TrimSpace(authHeader) != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := s.outboundHTTP.Do(req)
	if err != nil {
		return nil, apierrors.ServiceUnavailable("academic-service unreachable for teacher assignment check")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, apierrors.Forbidden("cannot verify teacher assignments")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, apierrors.ServiceUnavailable("academic-service teacher assignment lookup failed")
	}

	var rows []teacherAssignmentRow
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, apierrors.ServiceUnavailable("invalid response from academic-service")
	}
	return rows, nil
}

func (s *AttendanceService) assertTeacherAssignedToClassSubject(ctx context.Context, 
	roleName string,
	teacherUserID uuid.UUID,
	classID uuid.UUID,
	subjectID *uuid.UUID,
	authHeader string,
) error {
	if roleName == "super_admin" || roleName != "teacher" {
		return nil
	}
	if subjectID == nil {
		return apierrors.Forbidden("subject_id is required for teachers")
	}

	assignments, err := s.fetchTeacherAssignments(ctx, teacherUserID, authHeader)
	if err != nil {
		return err
	}

	classStr := classID.String()
	subjectStr := subjectID.String()
	for _, a := range assignments {
		if a.ClassID == classStr && a.SubjectID == subjectStr {
			return nil
		}
	}
	return apierrors.Forbidden("you are not assigned to this class and subject")
}

func (s *AttendanceService) enforceTeacherAttendanceQuery(ctx context.Context, 
	roleName string,
	teacherUserID uuid.UUID,
	classID, subjectID string,
	authHeader string,
) error {
	if roleName == "super_admin" || roleName != "teacher" {
		return nil
	}
	if strings.TrimSpace(classID) == "" || strings.TrimSpace(subjectID) == "" {
		return apierrors.Forbidden("teachers must filter by class_id and subject_id")
	}
	parsedClass, err := uuid.Parse(classID)
	if err != nil {
		return errors.New("invalid class_id")
	}
	parsedSubject, err := uuid.Parse(subjectID)
	if err != nil {
		return errors.New("invalid subject_id")
	}
	return s.assertTeacherAssignedToClassSubject(ctx, roleName, teacherUserID, parsedClass, &parsedSubject, authHeader)
}

func (s *AttendanceService) validateAuthUserInSchool(ctx context.Context, userID, schoolID uuid.UUID) error {
	if strings.TrimSpace(s.cfg.InternalServiceToken) == "" {
		return apierrors.ServiceUnavailable("user validation is not configured (set INTERNAL_SERVICE_TOKEN and USER_SERVICE_URL)")
	}
	base := strings.TrimRight(s.cfg.UserServiceURL, "/")
	url := fmt.Sprintf("%s/internal/users/%s?school_id=%s", base, userID.String(), schoolID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to build auth request: %w", err)
	}

	resp, err := s.userInternal.DoContext(ctx, req)
	if err != nil {
		return apierrors.ServiceUnavailable("user-service unreachable for user validation")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return apierrors.BadRequest("user not found")
	}
	if resp.StatusCode != http.StatusOK {
		return apierrors.ServiceUnavailable("user-service user validation failed")
	}

	var u authUserInternal
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return apierrors.ServiceUnavailable("invalid response from user-service")
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

func (s *AttendanceService) CreateAttendance(ctx context.Context, 
	req model.CreateAttendanceRequest,
	schoolID, teacherUserID uuid.UUID,
	roleName, authHeader string,
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

	if err := s.assertTeacherAssignedToClassSubject(ctx, roleName, teacherUserID, classID, subjectID, authHeader); err != nil {
		return nil, err
	}

	_, err = s.repo.GetAttendanceByComposite(schoolID, studentID, classID, sectionID, subjectID, attendanceDate)
	if err == nil {
		return nil, errors.New("attendance already marked for this student and date")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("create attendance: uniqueness check failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("student_id", studentID))
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
		log.Error("create attendance: database insert failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("student_id", studentID))
		return nil, fmt.Errorf("failed to create attendance: %w", err)
	}
	log.Info("attendance created",
		log.AddField("attendance_id", record.ID),
		log.AddField("school_id", schoolID),
		log.AddField("student_id", studentID),
		log.AddField("status", record.Status),
	)

	return record, nil
}

func (s *AttendanceService) GetAttendance(ctx context.Context, 
	schoolID uuid.UUID,
	query model.AttendanceQuery,
	requestingUserID uuid.UUID,
	roleName, authHeader string,
) (*model.AttendanceListResponse, error) {
	params := pagination.Params{Page: query.Page, Limit: query.Limit}
	pagination.Normalize(&params, pagination.Options{})
	query.Page = params.Page
	query.Limit = params.Limit

	if strings.TrimSpace(query.Date) != "" {
		if _, err := time.Parse("2006-01-02", query.Date); err != nil {
			return nil, errors.New("invalid date format, use YYYY-MM-DD")
		}
	}

	if err := s.enforceTeacherAttendanceQuery(ctx, roleName, requestingUserID, query.ClassID, query.SubjectID, authHeader); err != nil {
		return nil, err
	}

	records, total, err := s.repo.GetAttendance(schoolID, query)
	if err != nil {
		log.Error("list attendance: database query failed", log.Err(err), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to fetch attendance: %w", err)
	}

	return &model.AttendanceListResponse{
		Attendance: records,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
	}, nil
}

func (s *AttendanceService) UpdateAttendance(ctx context.Context, 
	id uuid.UUID,
	req model.UpdateAttendanceRequest,
	schoolID, requestingUserID uuid.UUID,
	roleName, authHeader string,
) (*model.Attendance, error) {
	record, err := s.repo.GetAttendanceByIDAndSchool(id, schoolID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierrors.NotFound("attendance not found")
		}
		log.Error("update attendance: database fetch failed", log.Err(err), log.AddField("attendance_id", id), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to fetch attendance: %w", err)
	}

	if roleName == "teacher" {
		if err := s.assertTeacherAssignedToClassSubject(ctx, roleName, requestingUserID, record.ClassID, record.SubjectID, authHeader); err != nil {
			return nil, err
		}
	} else if roleName != "super_admin" && record.TeacherUserID != requestingUserID {
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
		log.Error("update attendance: database update failed", log.Err(err), log.AddField("attendance_id", id), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to update attendance: %w", err)
	}
	log.Info("attendance updated", log.AddField("attendance_id", record.ID), log.AddField("school_id", schoolID))

	return record, nil
}

func (s *AttendanceService) BulkCreateAttendance(ctx context.Context, 
	req model.BulkCreateAttendanceRequest,
	schoolID, teacherUserID uuid.UUID,
	roleName, authHeader string,
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

	if err := s.assertTeacherAssignedToClassSubject(ctx, roleName, teacherUserID, classID, subjectID, authHeader); err != nil {
		return nil, err
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

	if resp.Created > 0 {
		log.Info("attendance bulk created",
			log.AddField("school_id", schoolID),
			log.AddField("created", resp.Created),
			log.AddField("skipped", resp.Skipped),
		)
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

func (s *AttendanceService) CreateTeacherAttendance(ctx context.Context, 
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

	if err := s.validateAuthUserInSchool(ctx, targetTeacherID, schoolID); err != nil {
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
		log.Error("create teacher attendance: uniqueness check failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("teacher_user_id", targetTeacherID))
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
		log.Error("create teacher attendance: database insert failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("teacher_user_id", targetTeacherID))
		return nil, fmt.Errorf("failed to create teacher attendance: %w", err)
	}
	log.Info("teacher attendance created",
		log.AddField("teacher_attendance_id", record.ID),
		log.AddField("school_id", schoolID),
		log.AddField("teacher_user_id", targetTeacherID),
		log.AddField("status", record.Status),
	)

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

func (s *AttendanceService) BulkCreateTeacherAttendance(ctx context.Context, 
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

		if err := s.validateAuthUserInSchool(ctx, teacherID, schoolID); err != nil {
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

	if resp.Created > 0 {
		log.Info("teacher attendance bulk created",
			log.AddField("school_id", schoolID),
			log.AddField("created", resp.Created),
			log.AddField("skipped", resp.Skipped),
		)
	}

	return resp, nil
}

func (s *AttendanceService) GetTeacherAttendance(
	schoolID, currentUserID uuid.UUID,
	roleName string,
	perms []string,
	query model.TeacherAttendanceQuery,
) (*model.TeacherAttendanceListResponse, error) {
	params := pagination.Params{Page: query.Page, Limit: query.Limit}
	pagination.Normalize(&params, pagination.Options{})
	query.Page = params.Page
	query.Limit = params.Limit

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
		log.Error("list teacher attendance: database query failed", log.Err(err), log.AddField("school_id", schoolID))
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
		log.Error("update teacher attendance: database fetch failed", log.Err(err), log.AddField("teacher_attendance_id", id), log.AddField("school_id", schoolID))
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
		log.Error("update teacher attendance: database update failed", log.Err(err), log.AddField("teacher_attendance_id", id), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to update teacher attendance: %w", err)
	}
	log.Info("teacher attendance updated", log.AddField("teacher_attendance_id", record.ID), log.AddField("school_id", schoolID))

	return record, nil
}

// GetAttendanceStats calculates attendance percentages for students.
func (s *AttendanceService) GetAttendanceStats(ctx context.Context, 
	schoolID uuid.UUID,
	query model.AttendanceStatsQuery,
	requestingUserID uuid.UUID,
	roleName, authHeader string,
) (*model.AttendanceStatsResponse, error) {
	if err := s.enforceTeacherAttendanceQuery(ctx, roleName, requestingUserID, query.ClassID, query.SubjectID, authHeader); err != nil {
		return nil, err
	}

	startDate, endDate, err := s.parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		return nil, err
	}

	counts, err := s.repo.GetAttendanceStats(schoolID, query, startDate, endDate)
	if err != nil {
		log.Error("get attendance stats: database query failed", log.Err(err), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to fetch attendance stats: %w", err)
	}

	statsMap := make(map[string]*model.AttendanceStats)
	for _, c := range counts {
		sid := c.StudentID
		if _, ok := statsMap[sid]; !ok {
			statsMap[sid] = &model.AttendanceStats{StudentID: sid, ClassID: query.ClassID}
		}
		st := statsMap[sid]
		switch c.Status {
		case "present":
			st.PresentDays = c.Count
		case "absent":
			st.AbsentDays = c.Count
		case "late":
			st.LateDays = c.Count
		case "excused":
			st.ExcusedDays = c.Count
		}
	}

	var stats []model.AttendanceStats
	for _, st := range statsMap {
		st.TotalDays = st.PresentDays + st.AbsentDays + st.LateDays + st.ExcusedDays
		if st.TotalDays > 0 {
			st.AttendanceRate = float64(st.PresentDays+st.LateDays+st.ExcusedDays) / float64(st.TotalDays) * 100
		}
		stats = append(stats, *st)
	}

	return &model.AttendanceStatsResponse{
		Stats:     stats,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
	}, nil
}

// GetTeacherAttendanceStats calculates attendance percentages for teachers.
func (s *AttendanceService) GetTeacherAttendanceStats(
	schoolID, currentUserID uuid.UUID,
	roleName string,
	perms []string,
	query model.TeacherAttendanceStatsQuery,
) (*model.TeacherAttendanceStatsResponse, error) {
	if roleName != "super_admin" && !hasPerm(perms, "view_teacher_attendance") && !hasPerm(perms, "mark_teacher_attendance") {
		if !hasPerm(perms, "mark_own_teacher_attendance") {
			return nil, apierrors.Forbidden("cannot view teacher attendance stats")
		}
		query.TeacherUserID = currentUserID.String()
	}

	startDate, endDate, err := s.parseDateRange(query.StartDate, query.EndDate)
	if err != nil {
		return nil, err
	}

	counts, err := s.repo.GetTeacherAttendanceStats(schoolID, query, startDate, endDate)
	if err != nil {
		log.Error("get teacher attendance stats: database query failed", log.Err(err), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to fetch teacher attendance stats: %w", err)
	}

	statsMap := make(map[string]*model.TeacherAttendanceStats)
	for _, c := range counts {
		tid := c.TeacherUserID
		if _, ok := statsMap[tid]; !ok {
			statsMap[tid] = &model.TeacherAttendanceStats{TeacherUserID: tid}
		}
		st := statsMap[tid]
		switch c.Status {
		case "present":
			st.PresentDays = c.Count
		case "absent":
			st.AbsentDays = c.Count
		case "late":
			st.LateDays = c.Count
		case "excused":
			st.ExcusedDays = c.Count
		}
	}

	var stats []model.TeacherAttendanceStats
	for _, st := range statsMap {
		st.TotalDays = st.PresentDays + st.AbsentDays + st.LateDays + st.ExcusedDays
		if st.TotalDays > 0 {
			st.AttendanceRate = float64(st.PresentDays+st.LateDays+st.ExcusedDays) / float64(st.TotalDays) * 100
		}
		stats = append(stats, *st)
	}

	return &model.TeacherAttendanceStatsResponse{
		Stats:     stats,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
	}, nil
}

// parseDateRange parses start/end date strings; defaults to current month if empty.
func (s *AttendanceService) parseDateRange(startStr, endStr string) (time.Time, time.Time, error) {
	now := time.Now()
	var startDate, endDate time.Time
	var err error

	if strings.TrimSpace(startStr) != "" {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("invalid start_date format, use YYYY-MM-DD")
		}
	} else {
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if strings.TrimSpace(endStr) != "" {
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("invalid end_date format, use YYYY-MM-DD")
		}
	} else {
		endDate = now
	}

	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, errors.New("end_date cannot be before start_date")
	}

	return startDate, endDate, nil
}
