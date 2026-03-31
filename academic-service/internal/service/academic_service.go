package service

import (
	"errors"
	"fmt"

	"github.com/avaneeshravat/school-management/academic-service/internal/model"
	"github.com/avaneeshravat/school-management/academic-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AcademicService struct {
	repo *repository.AcademicRepository
}

func NewAcademicService(repo *repository.AcademicRepository) *AcademicService {
	return &AcademicService{repo: repo}
}

func (s *AcademicService) CreateClass(req model.CreateClassRequest, schoolID uuid.UUID) (*model.Class, error) {
	_, err := s.repo.GetClassByName(schoolID, req.Name)
	if err == nil {
		return nil, errors.New("class with this name already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to validate class uniqueness: %w", err)
	}

	class := &model.Class{
		SchoolID:    schoolID,
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.repo.CreateClass(class); err != nil {
		return nil, fmt.Errorf("failed to create class: %w", err)
	}
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
		return nil, fmt.Errorf("failed to validate section uniqueness: %w", err)
	}

	section := &model.Section{
		SchoolID: schoolID,
		ClassID:  classID,
		Name:     req.Name,
	}
	if err := s.repo.CreateSection(section); err != nil {
		return nil, fmt.Errorf("failed to create section: %w", err)
	}
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
		return nil, fmt.Errorf("failed to create subject: %w", err)
	}
	return subject, nil
}

func (s *AcademicService) GetClasses(schoolID uuid.UUID) ([]model.ClassWithChildren, error) {
	classes, err := s.repo.GetClassesBySchoolID(schoolID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch classes: %w", err)
	}

	resp := make([]model.ClassWithChildren, 0, len(classes))
	for _, class := range classes {
		sections, err := s.repo.GetSectionsByClassID(class.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch sections for class %s: %w", class.ID, err)
		}
		subjects, err := s.repo.GetSubjectsByClassID(class.ID)
		if err != nil {
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
