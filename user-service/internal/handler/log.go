package handler

import (
	"net/http"

	"github.com/Archiit19/School-management/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func requestLogger(c *gin.Context) logger.Logger {
	var fields []logger.Field
	if id, ok := c.Get("request_id"); ok {
		fields = append(fields, logger.Any("request_id", id))
	}
	if uid, ok := c.Get("user_id"); ok {
		fields = append(fields, logger.Any("actor_user_id", uid))
	}
	if sid, ok := c.Get("school_id"); ok {
		fields = append(fields, logger.Any("school_id", sid))
	}
	return logger.With(fields...)
}

func logBindError(c *gin.Context, err error) {
	requestLogger(c).Warn("invalid request payload", logger.Err(err))
}

func logServiceError(c *gin.Context, status int, msg string, err error, fields ...logger.Field) {
	all := append([]logger.Field{logger.Err(err)}, fields...)
	if status >= http.StatusInternalServerError {
		requestLogger(c).Error(msg, all...)
		return
	}
	requestLogger(c).Warn(msg, all...)
}

func logUserID(id uuid.UUID) logger.Field {
	return logger.String("user_id", id.String())
}

func logSchoolID(id uuid.UUID) logger.Field {
	return logger.String("school_id", id.String())
}

func uuidField(key string, id uuid.UUID) logger.Field {
	return logger.String(key, id.String())
}
