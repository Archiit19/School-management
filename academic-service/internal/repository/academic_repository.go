package repository

import (
	"github.com/avaneeshravat/school-management/academic-service/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AcademicRepository struct {
	db *gorm.DB
}

func NewAcademicRepository(db *gorm.DB) *AcademicRepository {
	return &AcademicRepository{db: db}
}

func (r *AcademicRepository) CreateClass(class *model.Class) error {
	return r.db.Create(class).Error
}

func (r *AcademicRepository) GetClassByName(schoolID uuid.UUID, name string) (*model.Class, error) {
	var class model.Class
	err := r.db.Where("school_id = ? AND name = ?", schoolID, name).First(&class).Error
	return &class, err
}

func (r *AcademicRepository) GetClassByIDAndSchool(classID, schoolID uuid.UUID) (*model.Class, error) {
	var class model.Class
	err := r.db.Where("id = ? AND school_id = ?", classID, schoolID).First(&class).Error
	return &class, err
}

func (r *AcademicRepository) GetClassesBySchoolID(schoolID uuid.UUID) ([]model.Class, error) {
	var classes []model.Class
	err := r.db.Where("school_id = ?", schoolID).Order("created_at asc").Find(&classes).Error
	return classes, err
}

func (r *AcademicRepository) CreateSection(section *model.Section) error {
	return r.db.Create(section).Error
}

func (r *AcademicRepository) GetSectionByClassAndName(classID uuid.UUID, name string) (*model.Section, error) {
	var section model.Section
	err := r.db.Where("class_id = ? AND name = ?", classID, name).First(&section).Error
	return &section, err
}

func (r *AcademicRepository) GetSectionByIDAndSchool(sectionID, schoolID uuid.UUID) (*model.Section, error) {
	var section model.Section
	err := r.db.Where("id = ? AND school_id = ?", sectionID, schoolID).First(&section).Error
	return &section, err
}

func (r *AcademicRepository) GetSectionsByClassID(classID uuid.UUID) ([]model.Section, error) {
	var sections []model.Section
	err := r.db.Where("class_id = ?", classID).Order("created_at asc").Find(&sections).Error
	return sections, err
}

func (r *AcademicRepository) CreateSubject(subject *model.Subject) error {
	return r.db.Create(subject).Error
}

func (r *AcademicRepository) GetSubjectByNameAndScope(
	schoolID, classID uuid.UUID,
	sectionID *uuid.UUID,
	name string,
) (*model.Subject, error) {
	var subject model.Subject
	query := r.db.Where("school_id = ? AND class_id = ? AND name = ?", schoolID, classID, name)
	if sectionID == nil {
		query = query.Where("section_id IS NULL")
	} else {
		query = query.Where("section_id = ?", *sectionID)
	}
	err := query.First(&subject).Error
	return &subject, err
}

func (r *AcademicRepository) GetSubjectsByClassID(classID uuid.UUID) ([]model.Subject, error) {
	var subjects []model.Subject
	err := r.db.Where("class_id = ?", classID).Order("created_at asc").Find(&subjects).Error
	return subjects, err
}

func (r *AcademicRepository) GetSubjectByIDAndSchool(subjectID, schoolID uuid.UUID) (*model.Subject, error) {
	var subject model.Subject
	err := r.db.Where("id = ? AND school_id = ?", subjectID, schoolID).First(&subject).Error
	return &subject, err
}

func (r *AcademicRepository) CreateTeacherAssignment(assignment *model.TeacherAssignment) error {
	return r.db.Create(assignment).Error
}

func (r *AcademicRepository) GetTeacherAssignmentByComposite(
	schoolID, teacherUserID, classID, subjectID uuid.UUID,
) (*model.TeacherAssignment, error) {
	var assignment model.TeacherAssignment
	err := r.db.
		Where(
			"school_id = ? AND teacher_user_id = ? AND class_id = ? AND subject_id = ?",
			schoolID, teacherUserID, classID, subjectID,
		).
		First(&assignment).Error
	return &assignment, err
}

func (r *AcademicRepository) GetTeacherAssignments(
	schoolID uuid.UUID,
	query model.TeacherAssignmentQuery,
) ([]model.TeacherAssignment, error) {
	var assignments []model.TeacherAssignment
	q := r.db.Where("school_id = ?", schoolID)

	if query.TeacherUserID != "" {
		q = q.Where("teacher_user_id = ?", query.TeacherUserID)
	}
	if query.ClassID != "" {
		q = q.Where("class_id = ?", query.ClassID)
	}
	if query.SubjectID != "" {
		q = q.Where("subject_id = ?", query.SubjectID)
	}

	err := q.Order("created_at desc").Find(&assignments).Error
	return assignments, err
}

func (r *AcademicRepository) CreateAssignment(assignment *model.Assignment) error {
	return r.db.Create(assignment).Error
}

func (r *AcademicRepository) GetAssignments(
	schoolID uuid.UUID,
	query model.AssignmentQuery,
) ([]model.Assignment, error) {
	var assignments []model.Assignment
	q := r.db.Where("school_id = ?", schoolID)
	if query.ClassID != "" {
		q = q.Where("class_id = ?", query.ClassID)
	}
	if query.SubjectID != "" {
		q = q.Where("subject_id = ?", query.SubjectID)
	}
	if query.TeacherID != "" {
		q = q.Where("teacher_user_id = ?", query.TeacherID)
	}

	err := q.Order("created_at desc").Find(&assignments).Error
	return assignments, err
}

func (r *AcademicRepository) GetAssignmentByIDAndSchool(
	assignmentID, schoolID uuid.UUID,
) (*model.Assignment, error) {
	var assignment model.Assignment
	err := r.db.Where("id = ? AND school_id = ?", assignmentID, schoolID).First(&assignment).Error
	return &assignment, err
}

func (r *AcademicRepository) GetSubmissionByComposite(
	schoolID, assignmentID, studentID uuid.UUID,
) (*model.Submission, error) {
	var submission model.Submission
	err := r.db.
		Where("school_id = ? AND assignment_id = ? AND student_id = ?", schoolID, assignmentID, studentID).
		First(&submission).Error
	return &submission, err
}

func (r *AcademicRepository) CreateSubmission(submission *model.Submission) error {
	return r.db.Create(submission).Error
}
