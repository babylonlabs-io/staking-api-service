package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	testmock "github.com/babylonlabs-io/staking-api-service/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	healthCheckPath = "/healthcheck"
)

func TestHealthCheck(t *testing.T) {
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	url := testServer.Server.URL + healthCheckPath

	// Make a GET request to the health check endpoint
	resp, err := http.Get(url)
	assert.NoError(t, err, "making GET request to health check endpoint should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected HTTP 200 OK status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	var responseBody map[string]string
	err = json.Unmarshal(bodyBytes, &responseBody)
	assert.NoError(t, err, "unmarshalling response body should not fail")

	// Check that the response body is as expected
	assert.Equal(t, "Server is up and running", responseBody["data"], "expected response body to match")
}

// Test the db connection error case
func TestHealthCheckDBError(t *testing.T) {
	mockDB := new(testmock.DBClient)
	mockDB.On("Ping", mock.Anything).Return(io.EOF) // Expect db error

	testServer := setupTestServer(t, &TestServerDependency{MockDbClient: mockDB})

	defer testServer.Close()

	url := testServer.Server.URL + healthCheckPath

	// Make a GET request to the health check endpoint
	resp, err := http.Get(url)
	assert.NoError(t, err, "making GET request to health check endpoint should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "expected HTTP 500 Internal Server Error status")

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading response body should not fail")

	// Convert the response body to a string
	responseBody := string(bodyBytes)

	assert.Equal(t, "{\"errorCode\":\"INTERNAL_SERVICE_ERROR\",\"message\":\"Internal service error\"}", responseBody, "expected response body to match")
}

func TestOptionsRequest(t *testing.T) {
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	url := testServer.Server.URL + healthCheckPath

	// Make a OPTION request to the health check endpoint
	client := &http.Client{}
	req, err := http.NewRequest("OPTIONS", url, nil)
	assert.NoError(t, err, "making OPTION request to health check endpoint should not fail")
	req.Header.Add("Origin", "https://dashboard.testnet3.babylonlabs-io.io")
	req.Header.Add("Access-Control-Request-Headers", "Content-Type")
	req.Header.Add("Access-Control-Request-Method", "GET")

	// Send the request
	resp, err := client.Do(req)
	assert.NoError(t, err, "making OPTION request to polygon address check endpoint should not fail")
	defer resp.Body.Close()

	// Check that the status code is HTTP 204
	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "expected HTTP 204 OK status")
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"), "expected Access-Control-Allow-Origin to be *")
	assert.Equal(t, "GET", resp.Header.Get("Access-Control-Allow-Methods"), "expected Access-Control-Allow-Methods to be GET")
}

func TestSecurityHeaders(t *testing.T) {
	// Create a test server with the SecurityHeadersMiddleware
	testServer := setupTestServer(t, nil)
	defer testServer.Close()

	// Test the default CSP by making a request to a non-swagger endpoint
	url := testServer.Server.URL + healthCheckPath

	resp, err := http.Get(url)
	assert.NoError(t, err, "making GET request to /somepath should not fail")
	defer resp.Body.Close()

	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"), "expected X-Content-Type-Options to be nosniff")
	assert.Equal(t, "1; mode=block", resp.Header.Get("X-Xss-Protection"), "expected X-Xss-Protection to be 1; mode=block")
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"), "expected X-Frame-Options to be DENY")
	assert.Equal(t, "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; object-src 'none'; frame-ancestors 'self'; form-action 'self'; block-all-mixed-content; base-uri 'self';", resp.Header.Get("Content-Security-Policy"), "expected default Content-Security-Policy")
	assert.Equal(t, "strict-origin-when-cross-origin", resp.Header.Get("Referrer-Policy"), "expected Referrer-Policy to be strict-origin-when-cross-origin")

	// Test the Swagger CSP by making a request to the /swagger/* endpoint
	swaggerURL := testServer.Server.URL + "/swagger/index.html"
	resp, err = http.Get(swaggerURL)
	assert.NoError(t, err, "making GET request to /swagger/index.html should not fail")
	defer resp.Body.Close()

	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"), "expected X-Content-Type-Options to be nosniff")
	assert.Equal(t, "1; mode=block", resp.Header.Get("X-Xss-Protection"), "expected X-Xss-Protection to be 1; mode=block")
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"), "expected X-Frame-Options to be DENY")
	assert.Equal(t, "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com https://stackpath.bootstrap.com; style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com https://stackpath.bootstrap.com; img-src 'self' data: https://cdnjs.cloudflare.com https://stackpath.bootstrap.com; font-src 'self' https://cdnjs.cloudflare.com https://stackpath.bootstrap.com; object-src 'none'; frame-ancestors 'self'; form-action 'self'; block-all-mixed-content; base-uri 'self';", resp.Header.Get("Content-Security-Policy"), "expected Swagger Content-Security-Policy")
	assert.Equal(t, "strict-origin-when-cross-origin", resp.Header.Get("Referrer-Policy"), "expected Referrer-Policy to be strict-origin-when-cross-origin")
}
