package governor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Address(t *testing.T) {
	cfg := &Config{
		Host: "localhost",
		Port: 8080,
	}

	assert.Equal(t, "localhost:8080", cfg.Address())
}

func TestConfig_SetDefault(t *testing.T) {
	cfg := &Config{
		Host: "localhost",
		Port: 8080,
	}

	cfg.SetDefault()
	assert.Equal(t, "localhost", cfg.Host)
}

func TestServerInfo(t *testing.T) {
	info := ServerInfo{
		Address: "localhost:8080",
		Scheme:  "http",
		Attr:    map[string]string{"weight": "100"},
	}

	assert.Equal(t, "localhost:8080", info.Address)
	assert.Equal(t, "http", info.Scheme)
	assert.Equal(t, map[string]string{"weight": "100"}, info.Attr)
}

func TestNewServer(t *testing.T) {
	// Mock config
	originalConfig := config.Get(config.KeyGovernor)
	defer func() {
		// Restore original config if needed
		_ = config.Set(config.KeyGovernor, originalConfig)
	}()

	// Set test config
	testConfig := map[string]interface{}{
		"host": "127.0.0.1",
		"port": 0, // Use random port
	}
	_ = config.Set(config.KeyGovernor, testConfig)

	server := NewServer()
	require.NotNil(t, server)
	assert.NotNil(t, server.Server)
	assert.NotNil(t, server.listener)
	assert.NotNil(t, server.Config)
	assert.Equal(t, "http", server.info.Scheme)
	assert.NotEmpty(t, server.info.Address)

	// Clean up
	_ = server.Stop()
}

func TestServer_ServeAndStop(t *testing.T) {
	// Mock config
	testConfig := map[string]interface{}{
		"host": "127.0.0.1",
		"port": 0, // Use random port
	}
	_ = config.Set(config.KeyGovernor, testConfig)

	server := NewServer()
	require.NotNil(t, server)

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve()
	}()

	// Wait for server to start
	time.Sleep(time.Millisecond * 100)

	// Test that server is running
	resp, err := http.Get("http://" + server.info.Address + "/")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Stop server
	err = server.Stop()
	require.NoError(t, err)

	// Wait for serve to complete
	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second * 2):
		t.Fatal("Server did not stop within timeout")
	}
}

func TestServer_Info(t *testing.T) {
	testConfig := map[string]interface{}{
		"host": "127.0.0.1",
		"port": 0,
	}
	_ = config.Set(config.KeyGovernor, testConfig)

	server := NewServer()
	require.NotNil(t, server)

	info := server.Info()
	assert.Equal(t, server.info.Address, info.Address)
	assert.Equal(t, server.info.Scheme, info.Scheme)
	assert.Equal(t, server.info.Attr, info.Attr)

	_ = server.Stop()
}

func TestRespErr(t *testing.T) {
	w := httptest.NewRecorder()
	testErr := assert.AnError

	respErr(w, http.StatusBadRequest, testErr)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, testErr.Error(), response.Msg)
}

func TestRespSuccess(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		pretty   string
		expected string
	}{
		{
			name:     "simple data",
			data:     map[string]string{"key": "value"},
			pretty:   "",
			expected: `{"key":"value"}`,
		},
		{
			name:     "pretty data",
			data:     map[string]string{"key": "value"},
			pretty:   "true",
			expected: "{\n    \"key\": \"value\"\n}",
		},
		{
			name:     "nil data",
			data:     nil,
			pretty:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test?pretty="+tt.pretty, nil)

			respSuccess(w, r, tt.data)

			assert.Equal(t, http.StatusOK, w.Code)

			if tt.expected != "" {
				body := strings.TrimSpace(w.Body.String())
				assert.Equal(t, tt.expected, body)
			}
		})
	}
}

func TestRespNoContent(t *testing.T) {
	w := httptest.NewRecorder()

	respNoContent(w)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, 0, w.Body.Len())
}

func TestSetConfig(t *testing.T) {
	w := httptest.NewRecorder()
	body := strings.NewReader(`{"keys": ["test.key"], "data": ["test.value"]}`)
	r := httptest.NewRequest("POST", "/configs", body)

	setConfig(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestSetConfig_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	body := strings.NewReader(`invalid json`)
	r := httptest.NewRequest("POST", "/configs", body)

	setConfig(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response.Msg, "invalid character")
}

func testGetConfig(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/configs", nil)

	getConfig(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var response json.RawMessage
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.NotEmpty(t, response)
}

func TestConfigHandle(t *testing.T) {
	tests := []struct {
		method string
		testFn func(t *testing.T)
	}{
		{
			method: "GET",
			testFn: testGetConfig,
		},
		{
			method: "POST",
			testFn: func(t *testing.T) {
				w := httptest.NewRecorder()
				body := strings.NewReader(`{"keys": ["test.key"], "data": ["test.value"]}`)
				r := httptest.NewRequest("POST", "/configs", body)

				configHandle(w, r)

				assert.Equal(t, http.StatusNoContent, w.Code)
			},
		},
		{
			method: "PUT",
			testFn: func(t *testing.T) {
				w := httptest.NewRecorder()
				body := strings.NewReader(`{"keys": ["test.key"], "data": ["test.value"]}`)
				r := httptest.NewRequest("PUT", "/configs", body)

				configHandle(w, r)

				assert.Equal(t, http.StatusNoContent, w.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			tt.testFn(t)
		})
	}
}

func TestEnvHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/env", nil)

	// Set a test environment variable
	originalValue := os.Getenv("TEST_ENV_VAR")
	os.Setenv("TEST_ENV_VAR", "test_value")
	defer func() {
		if originalValue != "" {
			os.Setenv("TEST_ENV_VAR", originalValue)
		} else {
			os.Unsetenv("TEST_ENV_VAR")
		}
	}()

	envHandle(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var envVars []string
	err := json.NewDecoder(w.Body).Decode(&envVars)
	require.NoError(t, err)

	// Check that our test environment variable is present
	found := false
	for _, env := range envVars {
		if strings.Contains(env, "TEST_ENV_VAR=test_value") {
			found = true
			break
		}
	}
	assert.True(t, found, "TEST_ENV_VAR should be present in environment variables")
}

func TestNewBuildInfoHandle(t *testing.T) {
	info := &debug.BuildInfo{
		Path: "test/path",
	}

	handler := newBuildInfoHandle(info)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/build_info", nil)

	handler(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var response debug.BuildInfo
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, info.Path, response.Path)
}

func TestHandleFunc(t *testing.T) {
	// Clear existing routes
	routes = []string{}
	DefaultServeMux = http.NewServeMux()

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}

	HandleFunc("/test", testHandler)

	assert.Contains(t, routes, "/test")

	// Test that the handler is properly registered
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	DefaultServeMux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}

func TestRoutesHandle(t *testing.T) {
	// Clear existing routes and add some test routes
	routes = []string{"/test1", "/test2"}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	routesHandle(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, routes, response)
}

func TestInit(t *testing.T) {
	// Clear existing routes
	routes = []string{}
	DefaultServeMux = http.NewServeMux()

	// Call Init to register default handlers
	Init()

	// Check that default routes are registered
	expectedRoutes := []string{
		"/routes",
		"/debug/pprof/",
		"/debug/pprof/cmdline",
		"/debug/pprof/profile",
		"/debug/pprof/symbol",
		"/debug/pprof/trace",
		"/env",
		"/configs",
	}

	for _, route := range expectedRoutes {
		assert.Contains(t, routes, route, "Route %s should be registered", route)
	}

	// Check build_info route if build info is available
	if _, ok := debug.ReadBuildInfo(); ok {
		assert.Contains(t, routes, "/build_info")
	}
}
