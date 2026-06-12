package handler

import (
	"net/http"

	log "github.com/Archiit19/School-management/pkg/logger"
	"github.com/gin-gonic/gin"
)

func requestLogger(c *gin.Context) log.Logger {
	var fields []log.Field
	if id, ok := c.Get("request_id"); ok {
		fields = append(fields, log.AddField("request_id", id))
	}
	if uid, ok := c.Get("user_id"); ok {
		fields = append(fields, log.AddField("actor_user_id", uid))
	}
	if sid, ok := c.Get("school_id"); ok {
		fields = append(fields, log.AddField("school_id", sid))
	}
	return log.With(fields...)
}

func logBindError(c *gin.Context, err error) {
	requestLogger(c).Warn("invalid request payload", log.Err(err))
}

func logServiceError(c *gin.Context, status int, msg string, err error, fields ...log.Field) {
	all := append([]log.Field{log.Err(err)}, fields...)
	if status >= http.StatusInternalServerError {
		requestLogger(c).Error(msg, all...)
		return
	}
	requestLogger(c).Warn(msg, all...)
}
