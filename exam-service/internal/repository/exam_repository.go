package repository

import (
	"github.com/avaneeshravat/school-management/exam-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExamRepository struct {
	db *gorm.DB
}

func NewExamRepository(db *gorm.DB) *ExamRepository {
	return &ExamRepository{db: db}
}

func (r *ExamRepository) CreateExam(exam *model.Exam) error {
	return r.db.Create(exam).Error
}

func (r *ExamRepository) GetExamByIDAndSchool(examID, schoolID uuid.UUID) (*model.Exam, error) {
	var exam model.Exam
	err := r.db.Where("id = ? AND school_id = ?", examID, schoolID).First(&exam).Error
	return &exam, err
}

func (r *ExamRepository) UpdateExam(exam *model.Exam) error {
	return r.db.Save(exam).Error
}

func (r *ExamRepository) GetMarkByExamAndStudent(examID, studentID uuid.UUID) (*model.Mark, error) {
	var mark model.Mark
	err := r.db.Where("exam_id = ? AND student_id = ?", examID, studentID).First(&mark).Error
	return &mark, err
}

func (r *ExamRepository) CreateMark(mark *model.Mark) error {
	return r.db.Create(mark).Error
}

func (r *ExamRepository) UpdateMark(mark *model.Mark) error {
	return r.db.Save(mark).Error
}

func (r *ExamRepository) GetResults(
	schoolID uuid.UUID,
	query model.ResultQuery,
	includeUnpublished bool,
) ([]model.ResultItem, error) {
	results := make([]model.ResultItem, 0)
	q := r.db.Table("marks m").
		Select(
			"m.exam_id, e.title as exam_title, e.exam_date, m.student_id, m.marks_obtained, e.total_marks, e.is_published as published",
		).
		Joins("join exams e on e.id = m.exam_id").
		Where("m.school_id = ?", schoolID)

	if !includeUnpublished {
		q = q.Where("e.is_published = ?", true)
	}
	if query.ExamID != "" {
		q = q.Where("m.exam_id = ?", query.ExamID)
	}
	if query.StudentID != "" {
		q = q.Where("m.student_id = ?", query.StudentID)
	}
	if query.ClassID != "" {
		q = q.Where("e.class_id = ?", query.ClassID)
	}

	err := q.Order("e.exam_date desc, m.created_at desc").Scan(&results).Error
	return results, err
}
