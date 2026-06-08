package service

// Platform admin permissions (JWT only — not stored in user-service).
var platformAdminPermissions = []string{
	"view_my_schools",
	"create_school",
	"select_school",
}

const platformAdminRole = "platform_admin"
