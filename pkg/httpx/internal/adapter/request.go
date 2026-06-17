package adapter

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// NewJSONRequest builds a request with JSON content type when body is non-nil.
func NewJSONRequest(baseURL, method, path string, body interface{}) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, JoinURL(baseURL, path), reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// JoinURL joins base URL with a path.
func JoinURL(baseURL, path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return baseURL + path
}
