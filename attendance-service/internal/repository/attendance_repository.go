package repository

import (
	"time"

	"github.com/avaneeshravat/school-management/attendance-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

func (r *AttendanceRepository) CreateAttendance(record *model.Attendance) error {
	return r.db.Create(record).Error
}

func (r *AttendanceRepository) GetAttendanceByComposite(
	schoolID, studentID, classID uuid.UUID,
	sectionID, subjectID *uuid.UUID,
	date time.Time,
) (*model.Attendance, error) {
	var record model.Attendance
	q := r.db.Where("school_id = ? AND student_id = ? AND class_id = ? AND date = ?", schoolID, studentID, classID, date)
	if sectionID == nil {
		q = q.Where("section_id IS NULL")
	} else {
		q = q.Where("section_id = ?", *sectionID)
	}
	if subjectID == nil {
		q = q.Where("subject_id IS NULL")
	} else {
		q = q.Where("subject_id = ?", *subjectID)
	}
	err := q.First(&record).Error
	return &record, err
}

func (r *AttendanceRepository) GetAttendanceByIDAndSchool(
	id, schoolID uuid.UUID,
) (*model.Attendance, error) {
	var record model.Attendance
	err := r.db.Where("id = ? AND school_id = ?", id, schoolID).First(&record).Error
	return &record, err
}

func (r *AttendanceRepository) UpdateAttendance(record *model.Attendance) error {
	return r.db.Save(record).Error
}

func (r *AttendanceRepository) GetAttendance(
	schoolID uuid.UUID,
	query model.AttendanceQuery,
) ([]model.Attendance, int64, error) {
	var records []model.Attendance
	var total int64

	q := r.db.Model(&model.Attendance{}).Where("school_id = ?", schoolID)

	if query.Date != "" {
		parsedDate, err := time.Parse("2006-01-02", query.Date)
		if err == nil {
			q = q.Where("date = ?", parsedDate)
		}
	}
	if query.StudentID != "" {
		q = q.Where("student_id = ?", query.StudentID)
	}
	if query.ClassID != "" {
		q = q.Where("class_id = ?", query.ClassID)
	}
	if query.SectionID != "" {
		q = q.Where("section_id = ?", query.SectionID)
	}
	if query.SubjectID != "" {
		q = q.Where("subject_id = ?", query.SubjectID)
	}
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.Limit
	err := q.Order("date desc, created_at desc").Offset(offset).Limit(query.Limit).Find(&records).Error
	return records, total, err
}
