package service

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Archiit19/School-management/pkg/logger"
	"github.com/Archiit19/School-management/pkg/pagination"
	"github.com/Archiit19/School-management/user-service/internal/config"
	"github.com/Archiit19/School-management/user-service/internal/model"
	"github.com/Archiit19/School-management/user-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	repo     *repository.UserRepository
	profiles *repository.ProfileRepository
	auth     *authClient
	school   *schoolClient
	academic *academicClient
}

func NewUserService(
	repo *repository.UserRepository,
	profiles *repository.ProfileRepository,
	cfg *config.Config,
) *UserService {
	return &UserService{
		repo:     repo,
		profiles: profiles,
		auth:     newAuthClient(cfg),
		school:   newSchoolClient(cfg),
		academic: newAcademicClient(cfg, &http.Client{Timeout: 8 * time.Second}),
	}
}

func (s *UserService) rollbackCreate(userID uuid.UUID, schoolID *uuid.UUID) {
	fields := []log.Field{log.AddField("user_id", userID)}
	if schoolID != nil {
		fields = append(fields, log.AddField("school_id", *schoolID))
	}
	log.Warn("rolling back user creation", fields...)

	if schoolID != nil {
		_ = s.academic.DeleteEnrollment(userID, *schoolID)
		_ = s.auth.RemoveUserRole(userID, *schoolID)
		_ = s.school.RemoveMember(*schoolID, userID)
	}
	_ = s.profiles.Delete(userID)
	_ = s.auth.DeleteUserAuth(userID)
	_ = s.repo.Delete(userID)
}

func (s *UserService) CreateProfileInternal(req model.CreateProfileInternalRequest) (*model.User, error) {
	if _, err := s.repo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("user with this email already exists")
	}
	user := &model.User{Name: req.Name, Email: req.Email, IsActive: true}
	if err := s.repo.Create(user); err != nil {
		log.Error("create profile internal: database insert failed", log.Err(err), log.AddField("email", req.Email))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	log.Info("profile created (internal)", log.AddField("user_id", user.ID), log.AddField("email", user.Email))
	return user, nil
}

func (s *UserService) CreateUser(req model.CreateUserRequest, schoolID uuid.UUID) (*model.User, error) {
	if _, err := s.repo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("user with this email already exists")
	}
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, errors.New("invalid role_id format")
	}
	roleName, err := s.auth.GetRoleByID(roleID)
	if err != nil || roleName == "" {
		return nil, errors.New("role not found — make sure the role_id is valid")
	}

	roleData := req.RoleData
	if roleData == nil {
		roleData = map[string]interface{}{}
	}

	fieldDefs, _ := s.auth.GetRoleFields(roleID)
	if err := validateRoleData(fieldDefs, roleData); err != nil {
		return nil, err
	}

	if strings.EqualFold(roleName, "student") {
		if err := s.enrichStudentRoleData(schoolID, roleData); err != nil {
			return nil, err
		}
	}
	if strings.EqualFold(roleName, "parent") {
		enrichParentRoleData(roleData)
	}

	user := &model.User{Name: req.Name, Email: req.Email, IsActive: true}
	if err := s.repo.Create(user); err != nil {
		log.Error("create user: database insert failed", log.Err(err), log.AddField("school_id", schoolID), log.AddField("email", req.Email))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.profiles.Save(user.ID, roleID, schoolID, roleData); err != nil {
		log.Error("create user: save role profile failed", log.Err(err), log.AddField("user_id", user.ID), log.AddField("school_id", schoolID))
		s.rollbackCreate(user.ID, nil)
		return nil, fmt.Errorf("failed to save role profile: %w", err)
	}

	if err := s.auth.SetCredential(user.ID, req.Password); err != nil {
		log.Error("create user: set credentials failed", log.Err(err), log.AddField("user_id", user.ID))
		s.rollbackCreate(user.ID, nil)
		return nil, fmt.Errorf("failed to set credentials: %w", err)
	}
	if err := s.school.AddMember(schoolID, user.ID); err != nil {
		log.Error("create user: add school member failed", log.Err(err), log.AddField("user_id", user.ID), log.AddField("school_id", schoolID))
		s.rollbackCreate(user.ID, &schoolID)
		return nil, fmt.Errorf("failed to link user to school: %w", err)
	}
	if err := s.auth.AssignUserRole(user.ID, schoolID, roleID); err != nil {
		log.Error("create user: assign role failed", log.Err(err), log.AddField("user_id", user.ID), log.AddField("school_id", schoolID), log.AddField("role_id", roleID))
		s.rollbackCreate(user.ID, &schoolID)
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}

	if strings.EqualFold(roleName, "student") {
		classID := fmt.Sprint(roleData["class_id"])
		sectionID := fmt.Sprint(roleData["section_id"])
		if err := s.academic.UpsertEnrollment(user.ID, schoolID, classID, sectionID); err != nil {
			log.Error("create user: enroll student failed", log.Err(err), log.AddField("user_id", user.ID), log.AddField("school_id", schoolID))
			s.rollbackCreate(user.ID, &schoolID)
			return nil, fmt.Errorf("failed to enroll student: %w", err)
		}
		parentID, err := parseParentUserID(roleData)
		if err != nil {
			s.rollbackCreate(user.ID, &schoolID)
			return nil, err
		}
		if err := s.profiles.AppendChild(parentID, user.ID); err != nil {
			log.Error("create user: link student to parent failed",
				append([]log.Field{log.Err(err)}, logParentChild(parentID, user.ID)...)...)
			s.rollbackCreate(user.ID, &schoolID)
			return nil, fmt.Errorf("failed to link student to parent: %w", err)
		}
		log.Info("student linked to parent", logParentChild(parentID, user.ID)...)
	}

	user.SchoolID = &schoolID
	user.RoleID = &roleID
	user.RoleName = roleName
	user.RoleData = roleData
	log.Info("user created",
		log.AddField("user_id", user.ID),
		log.AddField("school_id", schoolID),
		log.AddField("role_name", roleName),
		log.AddField("email", user.Email),
	)
	return user, nil
}

func validateRoleData(defs []fieldDefinition, data map[string]interface{}) error {
	for _, f := range defs {
		if !f.Required {
			continue
		}
		val, ok := data[f.Key]
		if f.Type == "list" {
			if !ok || val == nil {
				return fmt.Errorf("required field missing: %s", f.Key)
			}
			continue
		}
		if !ok || strings.TrimSpace(fmt.Sprint(val)) == "" {
			return fmt.Errorf("required field missing: %s", f.Key)
		}
	}
	return nil
}

func enrichParentRoleData(data map[string]interface{}) {
	if data == nil {
		return
	}
	if _, ok := data["children"]; !ok || data["children"] == nil {
		data["children"] = []interface{}{}
	}
}

func parseParentUserID(data map[string]interface{}) (uuid.UUID, error) {
	parentIDStr := strings.TrimSpace(fmt.Sprint(data["parent_user_id"]))
	if parentIDStr == "" || parentIDStr == "<nil>" {
		return uuid.Nil, errors.New("parent_user_id is required")
	}
	return uuid.Parse(parentIDStr)
}

func (s *UserService) enrichStudentRoleData(schoolID uuid.UUID, data map[string]interface{}) error {
	classID, ok := data["class_id"].(string)
	if !ok || classID == "" {
		if v, ok := data["class_id"]; ok {
			classID = fmt.Sprint(v)
		}
	}
	if classID == "" {
		return errors.New("class_id is required for student role")
	}

	if _, ok := data["student_code"]; !ok || strings.TrimSpace(fmt.Sprint(data["student_code"])) == "" {
		code, err := s.generateStudentCode(schoolID, classID, data)
		if err != nil {
			return err
		}
		data["student_code"] = code
	}
	data["admission_year"] = time.Now().Year()

	parentIDStr := strings.TrimSpace(fmt.Sprint(data["parent_user_id"]))
	if parentIDStr == "" || parentIDStr == "<nil>" {
		return errors.New("parent is required for student role")
	}
	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		return errors.New("invalid parent_user_id")
	}
	parentUser, err := s.validateParentUser(parentID, schoolID)
	if err != nil {
		return err
	}
	data["parent_user_id"] = parentID.String()
	data["parent_name"] = parentUser.Name
	return nil
}

func (s *UserService) validateParentUser(parentID, schoolID uuid.UUID) (*model.User, error) {
	if err := s.school.GetMembership(schoolID, parentID); err != nil {
		return nil, errors.New("parent user not found in this school")
	}
	parent, err := s.repo.GetByID(parentID)
	if err != nil {
		return nil, errors.New("parent user not found")
	}
	if !parent.IsActive {
		return nil, errors.New("parent user account is inactive")
	}
	ur, err := s.auth.GetUserRole(parentID, schoolID)
	if err != nil {
		return nil, errors.New("parent role not found for user")
	}
	if !strings.EqualFold(ur.RoleName, "parent") {
		return nil, errors.New("linked user must have the parent role")
	}
	return parent, nil
}

func (s *UserService) generateStudentCode(schoolID uuid.UUID, classID string, data map[string]interface{}) (string, error) {
	admissionYear := time.Now().Year()
	sectionID := ""
	if v, ok := data["section_id"]; ok {
		sectionID = fmt.Sprint(v)
	}
	classNum := "00"
	sectionLetter := "X"
	// Use simple prefix; full academic validation can be added via academic-service
	codePrefix := fmt.Sprintf("%04d%s%s", admissionYear, classNum, sectionLetter)
	_ = schoolID
	_ = classID
	_ = sectionID
	return fmt.Sprintf("%s%02d", codePrefix, 1), nil
}

func (s *UserService) GetUsers(schoolID uuid.UUID, query model.UserListQuery) (*model.UserListResponse, error) {
	params := pagination.Params{Page: query.Page, Limit: query.Limit}
	if query.IDs != "" {
		pagination.Normalize(&params, pagination.Options{MaxLimit: 200})
	} else {
		pagination.Normalize(&params, pagination.Options{MaxLimit: 100})
	}
	query.Page = params.Page
	query.Limit = params.Limit

	memberIDs, err := s.school.ListMemberUserIDs(schoolID)
	if err != nil {
		log.Error("list users: school members lookup failed", log.Err(err), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to list school members: %w", err)
	}

	if query.RoleID != "" {
		rid, err := uuid.Parse(query.RoleID)
		if err == nil {
			filtered := make([]uuid.UUID, 0)
			for _, uid := range memberIDs {
				ur, err := s.auth.GetUserRole(uid, schoolID)
				if err == nil && ur.RoleID == rid {
					filtered = append(filtered, uid)
				}
			}
			memberIDs = filtered
		}
	}

	if query.IDs != "" {
		want := parseUUIDList(query.IDs)
		if len(want) > 0 {
			allowed := make(map[uuid.UUID]struct{}, len(memberIDs))
			for _, id := range memberIDs {
				allowed[id] = struct{}{}
			}
			filtered := make([]uuid.UUID, 0, len(want))
			for _, id := range want {
				if _, ok := allowed[id]; ok {
					filtered = append(filtered, id)
				}
			}
			memberIDs = filtered
		}
	}

	profiles, _ := s.profiles.BatchGet(memberIDs)

	users, total, err := s.repo.GetByIDs(memberIDs, query)
	if err != nil {
		log.Error("list users: database query failed", log.Err(err), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	for i := range users {
		if ur, err := s.auth.GetUserRole(users[i].ID, schoolID); err == nil {
			users[i].RoleID = &ur.RoleID
			users[i].RoleName = ur.RoleName
		}
		users[i].SchoolID = &schoolID
		if p, ok := profiles[users[i].ID]; ok {
			users[i].RoleData = p.Data
		}
		s.mergeEnrollmentIntoRoleData(&users[i], schoolID)
	}

	return &model.UserListResponse{Users: users, Total: total, Page: query.Page, Limit: query.Limit}, nil
}

func (s *UserService) GetUserByID(id uuid.UUID, schoolID uuid.UUID) (*model.User, error) {
	if err := s.school.GetMembership(schoolID, id); err != nil {
		return nil, errors.New("user not found")
	}
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	if ur, err := s.auth.GetUserRole(id, schoolID); err == nil {
		user.RoleID = &ur.RoleID
		user.RoleName = ur.RoleName
	}
	user.SchoolID = &schoolID
	if p, err := s.profiles.Get(id); err == nil {
		user.RoleData = p.Data
	}
	s.mergeEnrollmentIntoRoleData(user, schoolID)
	return user, nil
}

func (s *UserService) mergeEnrollmentIntoRoleData(user *model.User, schoolID uuid.UUID) {
	if enrollment, err := s.academic.GetEnrollment(user.ID, schoolID); err == nil && enrollment != nil {
		if user.RoleData == nil {
			user.RoleData = map[string]interface{}{}
		}
		user.RoleData["class_id"] = enrollment.ClassID
		if enrollment.SectionID != nil {
			user.RoleData["section_id"] = *enrollment.SectionID
		}
	}
}

func (s *UserService) GetUserMe(userID, schoolID uuid.UUID) (*model.User, error) {
	user, err := s.GetUserByID(userID, schoolID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, errors.New("user account inactive")
	}
	return user, nil
}

func (s *UserService) GetUserForInternal(id uuid.UUID, schoolID *uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	if schoolID != nil && *schoolID != uuid.Nil {
		if ur, err := s.auth.GetUserRole(id, *schoolID); err == nil {
			user.SchoolID = schoolID
			user.RoleID = &ur.RoleID
			user.RoleName = ur.RoleName
		}
	} else {
		schools, err := s.school.ListMembershipsForUser(id)
		if err == nil && len(schools) == 1 {
			sid := schools[0]
			if ur, err := s.auth.GetUserRole(id, sid); err == nil {
				user.SchoolID = &sid
				user.RoleID = &ur.RoleID
				user.RoleName = ur.RoleName
			}
		}
	}
	if p, err := s.profiles.Get(id); err == nil {
		user.RoleData = p.Data
	}
	return user, nil
}

func (s *UserService) GetUserProfileInternal(id uuid.UUID) (map[string]interface{}, error) {
	p, err := s.profiles.Get(id)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{
		"user_id":   p.UserID,
		"role_id":   p.RoleID,
		"school_id": p.SchoolID,
	}
	for k, v := range p.Data {
		result[k] = v
	}
	schoolID, _ := uuid.Parse(p.SchoolID)
	if enrollment, err := s.academic.GetEnrollment(id, schoolID); err == nil && enrollment != nil {
		result["class_id"] = enrollment.ClassID
		if enrollment.SectionID != nil {
			result["section_id"] = *enrollment.SectionID
		}
	}
	return result, nil
}

func (s *UserService) UpdateUser(id uuid.UUID, req model.UpdateUserRequest, schoolID uuid.UUID) (*model.User, error) {
	if err := s.school.GetMembership(schoolID, id); err != nil {
		return nil, errors.New("user not found")
	}
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		existing, err := s.repo.GetByEmail(*req.Email)
		if err == nil && existing.ID != user.ID {
			return nil, errors.New("email already in use by another user")
		}
		user.Email = *req.Email
	}
	var roleID uuid.UUID
	if ur, err := s.auth.GetUserRole(id, schoolID); err == nil {
		roleID = ur.RoleID
		user.RoleID = &ur.RoleID
		user.RoleName = ur.RoleName
	}
	if req.RoleID != nil {
		parsed, err := uuid.Parse(*req.RoleID)
		if err != nil {
			return nil, errors.New("invalid role_id format")
		}
		roleName, err := s.auth.GetRoleByID(parsed)
		if err != nil || roleName == "" {
			return nil, errors.New("role not found")
		}
		if err := s.auth.UpdateUserRole(id, schoolID, parsed); err != nil {
			return nil, fmt.Errorf("failed to update role: %w", err)
		}
		roleID = parsed
		user.RoleID = &parsed
		user.RoleName = roleName
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if err := s.repo.Update(user); err != nil {
		log.Error("update user: database update failed", log.Err(err), log.AddField("user_id", id), log.AddField("school_id", schoolID))
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if req.RoleData != nil && roleID != uuid.Nil {
		existing, _ := s.profiles.Get(id)
		var oldParentID uuid.UUID
		if existing != nil && existing.Data != nil {
			oldParentID, _ = parseParentUserIDOptional(existing.Data)
		}
		data := req.RoleData
		if existing != nil && existing.Data != nil {
			for k, v := range existing.Data {
				if _, ok := data[k]; !ok {
					data[k] = v
				}
			}
		}
		fieldDefs, _ := s.auth.GetRoleFields(roleID)
		if err := validateRoleData(fieldDefs, data); err != nil {
			return nil, err
		}
		if strings.EqualFold(user.RoleName, "student") {
			if err := s.enrichStudentRoleData(schoolID, data); err != nil {
				return nil, err
			}
		}
		if strings.EqualFold(user.RoleName, "parent") {
			enrichParentRoleData(data)
		}
		if err := s.profiles.Save(id, roleID, schoolID, data); err != nil {
			log.Error("update user: save role profile failed", log.Err(err), log.AddField("user_id", id), log.AddField("school_id", schoolID))
			return nil, fmt.Errorf("failed to update role profile: %w", err)
		}
		user.RoleData = data
		if strings.EqualFold(user.RoleName, "student") {
			classID := fmt.Sprint(data["class_id"])
			sectionID := fmt.Sprint(data["section_id"])
			if classID != "" {
				if err := s.academic.UpsertEnrollment(id, schoolID, classID, sectionID); err != nil {
					log.Error("update user: enrollment update failed", log.Err(err), log.AddField("user_id", id), log.AddField("school_id", schoolID))
					return nil, fmt.Errorf("failed to update enrollment: %w", err)
				}
			}
			newParentID, _ := parseParentUserIDOptional(data)
			if oldParentID != newParentID {
				if oldParentID != uuid.Nil {
					_ = s.profiles.RemoveChild(oldParentID, id)
					log.Info("student unlinked from parent", logParentChild(oldParentID, id)...)
				}
				if newParentID != uuid.Nil {
					if err := s.profiles.AppendChild(newParentID, id); err != nil {
						log.Error("update user: link student to parent failed",
							append([]log.Field{log.Err(err)}, logParentChild(newParentID, id)...)...)
						return nil, fmt.Errorf("failed to link student to parent: %w", err)
					}
					log.Info("student linked to parent", logParentChild(newParentID, id)...)
				}
			}
		}
	} else if p, err := s.profiles.Get(id); err == nil {
		user.RoleData = p.Data
	}

	user.SchoolID = &schoolID
	log.Info("user updated", log.AddField("user_id", id), log.AddField("school_id", schoolID), log.AddField("role_name", user.RoleName))
	return user, nil
}

func (s *UserService) DeleteUser(id uuid.UUID, schoolID uuid.UUID, requestingUserID uuid.UUID) error {
	if id == requestingUserID {
		return errors.New("you cannot delete your own account")
	}
	if err := s.school.GetMembership(schoolID, id); err != nil {
		return errors.New("user not found")
	}
	if profile, err := s.profiles.Get(id); err == nil && profile != nil {
		if parentID, err := parseParentUserIDOptional(profile.Data); err == nil && parentID != uuid.Nil {
			_ = s.profiles.RemoveChild(parentID, id)
		}
	}
	if err := s.auth.RemoveUserRole(id, schoolID); err != nil {
		log.Error("delete user: remove role failed", log.Err(err), log.AddField("user_id", id), log.AddField("school_id", schoolID))
		return err
	}
	if err := s.school.RemoveMember(schoolID, id); err != nil {
		log.Error("delete user: remove school member failed", log.Err(err), log.AddField("user_id", id), log.AddField("school_id", schoolID))
		return err
	}
	_ = s.academic.DeleteEnrollment(id, schoolID)
	schools, err := s.school.ListMembershipsForUser(id)
	if err != nil {
		log.Error("delete user: list memberships failed", log.Err(err), log.AddField("user_id", id))
		return err
	}
	if len(schools) == 0 {
		_ = s.profiles.Delete(id)
		if err := s.auth.DeleteUserAuth(id); err != nil {
			log.Error("delete user: delete auth record failed", log.Err(err), log.AddField("user_id", id))
			return err
		}
		if err := s.repo.Delete(id); err != nil {
			log.Error("delete user: database delete failed", log.Err(err), log.AddField("user_id", id))
			return fmt.Errorf("failed to delete user: %w", err)
		}
		log.Info("user fully deleted", log.AddField("user_id", id), log.AddField("school_id", schoolID))
	} else {
		log.Info("user removed from school", log.AddField("user_id", id), log.AddField("school_id", schoolID), log.AddField("remaining_schools", len(schools)))
	}
	return nil
}

func (s *UserService) GetUserForInternalByEmail(email string) (*model.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) DeleteProfileInternal(id uuid.UUID) error {
	_ = s.profiles.Delete(id)
	if err := s.repo.Delete(id); err != nil {
		log.Error("delete profile internal: database delete failed", log.Err(err), log.AddField("user_id", id))
		return err
	}
	log.Info("profile deleted (internal)", log.AddField("user_id", id))
	return nil
}

func (s *UserService) UpdateProfileInternal(id uuid.UUID, name, email *string) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if name != nil {
		user.Name = *name
	}
	if email != nil && *email != user.Email {
		if _, err := s.repo.GetByEmail(*email); err == nil {
			return nil, errors.New("email already in use")
		}
		user.Email = *email
	}
	if err := s.repo.Update(user); err != nil {
		log.Error("update profile internal: database update failed", log.Err(err), log.AddField("user_id", id))
		return nil, err
	}
	log.Info("profile updated (internal)", log.AddField("user_id", id))
	return user, nil
}

func parseParentUserIDOptional(data map[string]interface{}) (uuid.UUID, error) {
	if data == nil {
		return uuid.Nil, nil
	}
	parentIDStr := strings.TrimSpace(fmt.Sprint(data["parent_user_id"]))
	if parentIDStr == "" || parentIDStr == "<nil>" {
		return uuid.Nil, nil
	}
	return uuid.Parse(parentIDStr)
}

func (s *UserService) ParentHasChild(parentID, childID uuid.UUID) (bool, error) {
	ok, err := s.profiles.HasChild(parentID, childID)
	if err != nil {
		log.Error("parent has-child check failed",
			append([]log.Field{log.Err(err)}, logParentChild(parentID, childID)...)...)
	}
	return ok, err
}

func (s *UserService) GetMyChildren(parentID, schoolID uuid.UUID) ([]model.User, error) {
	ur, err := s.auth.GetUserRole(parentID, schoolID)
	if err != nil {
		return nil, errors.New("parent role not found")
	}
	if !strings.EqualFold(ur.RoleName, "parent") {
		return nil, errors.New("only parent accounts can list children")
	}
	profile, err := s.profiles.Get(parentID)
	if err != nil {
		return nil, errors.New("parent profile not found")
	}
	childIDs := repository.ParseChildrenIDs(profile.Data["children"])
	children := make([]model.User, 0, len(childIDs))
	for _, idStr := range childIDs {
		childID, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		child, err := s.GetUserByID(childID, schoolID)
		if err != nil {
			continue
		}
		children = append(children, *child)
	}
	return children, nil
}

func (s *UserService) GetChildForParent(parentID, childID, schoolID uuid.UUID) (*model.User, error) {
	ok, err := s.profiles.HasChild(parentID, childID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("student is not linked to this parent account")
	}
	return s.GetUserByID(childID, schoolID)
}

func extractClassNumber(className string) string {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(className)
	if match == "" {
		return "00"
	}
	num, err := strconv.Atoi(match)
	if err != nil {
		return "00"
	}
	return fmt.Sprintf("%02d", num)
}

func extractSectionLetter(sectionName string) string {
	name := strings.ToUpper(strings.TrimSpace(sectionName))
	if name == "" {
		return "X"
	}
	if len(name) == 1 {
		return name
	}
	parts := strings.Fields(name)
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if len(last) == 1 {
			return last
		}
	}
	return string([]rune(name)[0])
}

func parseUUIDList(raw string) []uuid.UUID {
	parts := strings.Split(raw, ",")
	out := make([]uuid.UUID, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if id, err := uuid.Parse(p); err == nil {
			out = append(out, id)
		}
	}
	return out
}
