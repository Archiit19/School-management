package httputil

import (
	"errors"
	"net/http"

	"github.com/Archiit19/School-management/pkg/apierrors"
	"github.com/gin-gonic/gin"
)

// ErrorResponse is the standard JSON error body.
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteError maps known API errors to HTTP responses.
func WriteError(c *gin.Context, err error) {
	var he *apierrors.HTTP
	if errors.As(err, &he) {
		c.JSON(he.Status, ErrorResponse{Error: he.Error()})
		return
	}
	c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
}

// WriteMessage writes a simple message response.
func WriteMessage(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"message": msg})
}
