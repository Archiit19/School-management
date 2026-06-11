package service

import (
	"github.com/Archiit19/School-management/pkg/logger"
	"github.com/google/uuid"
)

func logUserID(id uuid.UUID) logger.Field {
	return logger.String("user_id", id.String())
}

func logSchoolID(id uuid.UUID) logger.Field {
	return logger.String("school_id", id.String())
}

func logRole(role string) logger.Field {
	return logger.String("role_name", role)
}

func logParentChild(parentID, childID uuid.UUID) []logger.Field {
	return []logger.Field{
		logger.String("parent_user_id", parentID.String()),
		logger.String("child_user_id", childID.String()),
	}
}
