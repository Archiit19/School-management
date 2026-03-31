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
