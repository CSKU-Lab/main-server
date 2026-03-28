//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HTTPResponse wraps an HTTP response for easier testing
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// JSON parses the response body as JSON
func (r *HTTPResponse) JSON(target interface{}) error {
	return json.Unmarshal(r.Body, target)
}

// String returns the response body as a string
func (r *HTTPResponse) String() string {
	return string(r.Body)
}

// RequestWithAuth makes an HTTP request with JWT authentication
func (s *TestSuite) RequestWithAuth(method, path string, body interface{}, token string) *HTTPResponse {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		s.Require().NoError(err)
		bodyReader = bytes.NewReader(jsonBody)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Cookie", fmt.Sprintf("access_token=%s", token))
	}

	resp, err := s.App.Test(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       bodyBytes,
		Headers:    resp.Header,
	}
}

// RequestWithoutAuth makes an HTTP request without authentication
func (s *TestSuite) RequestWithoutAuth(method, path string, body interface{}) *HTTPResponse {
	return s.RequestWithAuth(method, path, body, "")
}

// RequestWithFormData makes a multipart/form-data request with authentication
func (s *TestSuite) RequestWithFormData(method, path string, fields map[string]string, files map[string][]byte, token string) *HTTPResponse {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add fields
	for key, value := range fields {
		_ = writer.WriteField(key, value)
	}

	// Add files
	for filename, content := range files {
		part, err := writer.CreateFormFile(filename, filename)
		s.Require().NoError(err)
		_, err = part.Write(content)
		s.Require().NoError(err)
	}

	err := writer.Close()
	s.Require().NoError(err)

	req := httptest.NewRequest(method, path, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Cookie", fmt.Sprintf("access_token=%s", token))
	}

	resp, err := s.App.Test(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       bodyBytes,
		Headers:    resp.Header,
	}
}

// AssertStatusCode asserts that the response has the expected status code
func (s *TestSuite) AssertStatusCode(resp *HTTPResponse, expected int) {
	assert.Equal(s.T(), expected, resp.StatusCode,
		"Expected status code %d but got %d. Response: %s", expected, resp.StatusCode, resp.String())
}

// AssertSuccess asserts that the response is successful (2xx)
func (s *TestSuite) AssertSuccess(resp *HTTPResponse) {
	s.AssertStatusCode(resp, http.StatusOK)
}

// AssertCreated asserts that the response is 201 Created
func (s *TestSuite) AssertCreated(resp *HTTPResponse) {
	s.AssertStatusCode(resp, http.StatusCreated)
}

// AssertNoContent asserts that the response is 204 No Content
func (s *TestSuite) AssertNoContent(resp *HTTPResponse) {
	s.AssertStatusCode(resp, http.StatusNoContent)
}

// AssertBadRequest asserts that the response is 400 Bad Request
func (s *TestSuite) AssertBadRequest(resp *HTTPResponse) {
	s.AssertStatusCode(resp, http.StatusBadRequest)
}

// AssertUnauthorized asserts that the response is 401 Unauthorized
func (s *TestSuite) AssertUnauthorized(resp *HTTPResponse) {
	s.AssertStatusCode(resp, http.StatusUnauthorized)
}

// AssertForbidden asserts that the response is 403 Forbidden
func (s *TestSuite) AssertForbidden(resp *HTTPResponse) {
	s.AssertStatusCode(resp, http.StatusForbidden)
}

// AssertNotFound asserts that the response is 404 Not Found
func (s *TestSuite) AssertNotFound(resp *HTTPResponse) {
	s.AssertStatusCode(resp, http.StatusNotFound)
}

// AssertErrorResponse asserts that the response contains an error message
func (s *TestSuite) AssertErrorResponse(resp *HTTPResponse, expectedMessage string) {
	var result map[string]interface{}
	err := resp.JSON(&result)
	s.Require().NoError(err, "Failed to parse error response JSON")

	message, ok := result["message"].(string)
	if !ok {
		// Try "error" field
		message, ok = result["error"].(string)
	}

	if expectedMessage != "" {
		assert.Contains(s.T(), strings.ToLower(message), strings.ToLower(expectedMessage),
			"Expected error message to contain '%s' but got '%s'", expectedMessage, message)
	} else {
		assert.NotEmpty(s.T(), message, "Expected error message to be present")
	}
}

// ParseJSONResponse parses the response body into the target struct
func (s *TestSuite) ParseJSONResponse(resp *HTTPResponse, target interface{}) {
	err := resp.JSON(target)
	s.Require().NoError(err, "Failed to parse JSON response: %s", resp.String())
}

// GetResponseField extracts a field from a JSON response
func (s *TestSuite) GetResponseField(resp *HTTPResponse, field string) interface{} {
	var result map[string]interface{}
	s.ParseJSONResponse(resp, &result)
	return result[field]
}

// GetNestedResponseField extracts a nested field from a JSON response
func (s *TestSuite) GetNestedResponseField(resp *HTTPResponse, fields ...string) interface{} {
	var current interface{}
	s.ParseJSONResponse(resp, &current)

	for _, field := range fields {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[field]
		} else {
			s.T().Fatalf("Cannot access field '%s' in response", field)
		}
	}

	return current
}

// ExtractIDFromResponse extracts an ID from a response body
func (s *TestSuite) ExtractIDFromResponse(resp *HTTPResponse) string {
	var result map[string]interface{}
	s.ParseJSONResponse(resp, &result)

	// Try common ID field names
	idFields := []string{"id", "ID", "user_id", "course_id", "section_id", "lab_id", "submission_id"}
	for _, field := range idFields {
		if id, ok := result[field].(string); ok && id != "" {
			return id
		}
	}

	s.T().Fatal("Could not extract ID from response")
	return ""
}

// BuildURL builds a URL path with the API prefix
func BuildURL(path string) string {
	if strings.HasPrefix(path, "/api/v1") {
		return path
	}
	if strings.HasPrefix(path, "/") {
		return "/api/v1" + path
	}
	return "/api/v1/" + path
}

// NewTestApp creates a new Fiber app for testing
func NewTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			// Default error handler for tests
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})
}

// RequireStatus is a helper that requires a specific status code
func RequireStatus(t *testing.T, resp *HTTPResponse, expected int) {
	require.Equal(t, expected, resp.StatusCode,
		"Expected status code %d but got %d. Response: %s", expected, resp.StatusCode, resp.String())
}

// AssertContains asserts that the response body contains a string
func (s *TestSuite) AssertContains(resp *HTTPResponse, expected string) {
	assert.Contains(s.T(), resp.String(), expected)
}

// AssertJSONField asserts that a JSON field equals the expected value
func (s *TestSuite) AssertJSONField(resp *HTTPResponse, field string, expected interface{}) {
	value := s.GetResponseField(resp, field)
	assert.Equal(s.T(), expected, value)
}

// AssertPagination asserts that the response contains pagination info
func (s *TestSuite) AssertPagination(resp *HTTPResponse) {
	pagination := s.GetNestedResponseField(resp, "pagination")
	assert.NotNil(s.T(), pagination, "Response should contain pagination")
}

// AssertDataArray asserts that the response contains a data array
func (s *TestSuite) AssertDataArray(resp *HTTPResponse) []interface{} {
	data := s.GetNestedResponseField(resp, "data")
	if arr, ok := data.([]interface{}); ok {
		return arr
	}
	s.T().Fatal("Response data is not an array")
	return nil
}

// LoginAndGetToken performs login and returns the access token
func (s *TestSuite) LoginAndGetToken(username, password string) string {
	reqBody := map[string]string{
		"username": username,
		"password": password,
	}

	resp := s.RequestWithoutAuth("POST", BuildURL("/auth/sign-in/credential"), reqBody)
	s.AssertSuccess(resp)

	// Extract token from cookies or response body
	// In dev mode, tokens are returned in the response body
	var result map[string]interface{}
	s.ParseJSONResponse(resp, &result)

	// Check for token in response (dev mode)
	if token, ok := result["access_token"].(string); ok {
		return token
	}

	// Otherwise, extract from cookies
	cookies := resp.Headers["Set-Cookie"]
	for _, cookie := range cookies {
		if strings.HasPrefix(cookie, "access_token=") {
			parts := strings.Split(cookie, ";")
			for _, part := range parts {
				if strings.HasPrefix(part, "access_token=") {
					return strings.TrimPrefix(part, "access_token=")
				}
			}
		}
	}

	s.T().Fatal("Could not extract access token from response")
	return ""
}
