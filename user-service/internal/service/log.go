package service

import (
	log "github.com/Archiit19/School-management/pkg/logger"
	"github.com/google/uuid"
)

func logParentChild(parentID, childID uuid.UUID) []log.Field {
	return []log.Field{
		log.AddField("parent_user_id", parentID),
		log.AddField("child_user_id", childID),
	}
}
