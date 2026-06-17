package adapter

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ReadErrorBody reads response body for error messages.
func ReadErrorBody(resp *http.Response) string {
	b, _ := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(b))
}

// CheckStatus returns an error when status does not match expectStatus.
func CheckStatus(resp *http.Response, expectStatus int, context string) error {
	if resp.StatusCode == expectStatus {
		return nil
	}
	msg := ReadErrorBody(resp)
	if msg != "" {
		return fmt.Errorf("%s status %d: %s", context, resp.StatusCode, msg)
	}
	return fmt.Errorf("%s status %d", context, resp.StatusCode)
}

// CheckStatusAny returns an error when status is not one of the allowed codes.
func CheckStatusAny(resp *http.Response, context string, allowed ...int) error {
	for _, code := range allowed {
		if resp.StatusCode == code {
			return nil
		}
	}
	msg := ReadErrorBody(resp)
	if msg != "" {
		return fmt.Errorf("%s status %d: %s", context, resp.StatusCode, msg)
	}
	return fmt.Errorf("%s status %d", context, resp.StatusCode)
}
