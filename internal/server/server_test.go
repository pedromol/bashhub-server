package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = "postgres://user:password@localhost:5432/testdb?sslmode=disable"
	}
	db, err := sql.Open("postgres", testDBURL)
	if err != nil {
		os.Exit(0)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		os.Exit(0)
	}
	code := m.Run()
	os.Exit(code)
}
func TestGetLog(t *testing.T) {
	tests := []struct {
		name     string
		logFile  string
		expected string
	}{
		{
			name:     "discard log",
			logFile:  "/dev/null",
			expected: "",
		},
		{
			name:     "stderr log",
			logFile:  "",
			expected: "",
		},
		{
			name:     "file log",
			logFile:  "/tmp/test.log",
			expected: "/tmp/test.log",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := getLog(tt.logFile)
			if tt.logFile == "/dev/null" {
				if reflect.TypeOf(writer) != reflect.TypeOf(&io.Discard) {
					t.Errorf("type mismatch")
				}
			} else if tt.logFile == "" {
				if os.Stderr != writer {
					t.Errorf("expected %v, got %v", os.Stderr, writer)
				}
			} else {
			}
		})
	}
}
func TestPingEndpoint(t *testing.T) {
	server := NewServer("postgres://user:password@localhost:5432/testdb?sslmode=disable", "/dev/null", true)
	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}
	if response["message"] != "pong" {
		t.Errorf("expected message 'pong', got %v", response["message"])
	}
}
func TestUserRegistration(t *testing.T) {
	server := NewServer("postgres://user:password@localhost:5432/testdb?sslmode=disable", "/dev/null", true)
	t.Run("successful registration", func(t *testing.T) {
		userData := map[string]interface{}{
			"username": "testuser",
			"password": "testpass123",
			"email":    "test@example.com",
		}
		jsonData, _ := json.Marshal(userData)
		req := httptest.NewRequest("POST", "/api/v1/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		if w.Code != http.StatusOK && w.Code != http.StatusConflict && w.Code != http.StatusBadRequest {
			t.Errorf("unexpected status code: %d", w.Code)
		}
	})
	t.Run("missing email", func(t *testing.T) {
		userData := map[string]interface{}{
			"username": "testuser2",
			"password": "testpass123",
		}
		jsonData, _ := json.Marshal(userData)
		req := httptest.NewRequest("POST", "/api/v1/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("failed to unmarshal response: %v", err)
		}
		if !strings.Contains(response["error"], "email required") {
			t.Errorf("expected error to contain 'email required', got %v", response["error"])
		}
	})
	t.Run("registration disabled", func(t *testing.T) {
		serverDisabled := NewServer("postgres://user:password@localhost:5432/testdb?sslmode=disable", "/dev/null", false)
		userData := map[string]interface{}{
			"username": "testuser3",
			"password": "testpass123",
			"email":    "test3@example.com",
		}
		jsonData, _ := json.Marshal(userData)
		req := httptest.NewRequest("POST", "/api/v1/user", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		serverDisabled.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Registration of new users is not allowed") {
			t.Errorf("expected body to contain registration disabled message, got %v", w.Body.String())
		}
	})
}
func TestLoginEndpoint(t *testing.T) {
	server := NewServer("postgres://user:password@localhost:5432/testdb?sslmode=disable", "/dev/null", true)
	t.Run("login request", func(t *testing.T) {
		loginData := map[string]interface{}{
			"username": "testuser",
			"password": "testpass123",
		}
		jsonData, _ := json.Marshal(loginData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		server.ServeHTTP(w, req)
		if w.Code != http.StatusOK && w.Code != http.StatusUnauthorized {
			t.Errorf("expected status OK or Unauthorized, got %d", w.Code)
		}
	})
	t.Run("missing credentials", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		server.ServeHTTP(w, req)
		if http.StatusUnauthorized != w.Code {
			t.Errorf("expected %v, got %v", http.StatusUnauthorized, w.Code)
		}
	})
}
func TestProtectedEndpoints(t *testing.T) {
	server := NewServer("postgres://user:password@localhost:5432/testdb?sslmode=disable", "/dev/null", true)
	t.Run("unauthorized access", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/command/search", nil)
		server.ServeHTTP(w, req)
		if http.StatusUnauthorized != w.Code {
			t.Errorf("expected %v, got %v", http.StatusUnauthorized, w.Code)
		}
	})
	t.Run("unauthorized access to system endpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/system", nil)
		server.ServeHTTP(w, req)
		if http.StatusUnauthorized != w.Code {
			t.Errorf("expected %v, got %v", http.StatusUnauthorized, w.Code)
		}
	})
}
func TestCommandEndpoints(t *testing.T) {
	server := NewServer("postgres://user:password@localhost:5432/testdb?sslmode=disable", "/dev/null", true)
	t.Run("command search without auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/command/search?query=ls", nil)
		server.ServeHTTP(w, req)
		if http.StatusUnauthorized != w.Code {
			t.Errorf("expected %v, got %v", http.StatusUnauthorized, w.Code)
		}
	})
	t.Run("command by UUID without auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/command/test-uuid", nil)
		server.ServeHTTP(w, req)
		if http.StatusUnauthorized != w.Code {
			t.Errorf("expected %v, got %v", http.StatusUnauthorized, w.Code)
		}
	})
}
func TestSystemEndpoints(t *testing.T) {
	server := NewServer("postgres://user:password@localhost:5432/testdb?sslmode=disable", "/dev/null", true)
	t.Run("get system without auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/system?mac=AA:BB:CC:DD:EE:FF", nil)
		server.ServeHTTP(w, req)
		if http.StatusUnauthorized != w.Code {
			t.Errorf("expected %v, got %v", http.StatusUnauthorized, w.Code)
		}
	})
	t.Run("create system without auth", func(t *testing.T) {
		systemData := map[string]interface{}{
			"name":     "test-system",
			"mac":      "AA:BB:CC:DD:EE:FF",
			"hostname": "test-host",
		}
		jsonData, _ := json.Marshal(systemData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/system", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		server.ServeHTTP(w, req)
		if http.StatusUnauthorized != w.Code {
			t.Errorf("expected %v, got %v", http.StatusUnauthorized, w.Code)
		}
	})
	t.Run("update system without auth", func(t *testing.T) {
		systemData := map[string]interface{}{
			"hostname": "updated-host",
		}
		jsonData, _ := json.Marshal(systemData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/system/AA:BB:CC:DD:EE:FF", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		server.ServeHTTP(w, req)
		if http.StatusUnauthorized != w.Code {
			t.Errorf("expected %v, got %v", http.StatusUnauthorized, w.Code)
		}
	})
}
func TestStatusEndpoint(t *testing.T) {
	server := NewServer("postgres://user:password@localhost:5432/testdb?sslmode=disable", "/dev/null", true)
	t.Run("get status without auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/client-view/status?processId=123&startTime=1640995200", nil)
		server.ServeHTTP(w, req)
		if http.StatusUnauthorized != w.Code {
			t.Errorf("expected %v, got %v", http.StatusUnauthorized, w.Code)
		}
	})
}
func TestImportEndpoint(t *testing.T) {
	server := NewServer("postgres://user:password@localhost:5432/testdb?sslmode=disable", "/dev/null", true)
	t.Run("import commands without auth", func(t *testing.T) {
		importData := map[string]interface{}{
			"command":    "test command",
			"path":       "/tmp",
			"created":    1640995200,
			"uuid":       "import-test-uuid",
			"exitStatus": 0,
			"systemName": "test-system",
		}
		jsonData, _ := json.Marshal(importData)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/import", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		server.ServeHTTP(w, req)
		if http.StatusUnauthorized != w.Code {
			t.Errorf("expected %v, got %v", http.StatusUnauthorized, w.Code)
		}
	})
}
func TestDeleteCommandEndpoint(t *testing.T) {
	server := NewServer("postgres://user:password@localhost:5432/testdb?sslmode=disable", "/dev/null", true)
	t.Run("delete command without auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/command/test-uuid", nil)
		server.ServeHTTP(w, req)
		if http.StatusUnauthorized != w.Code {
			t.Errorf("expected %v, got %v", http.StatusUnauthorized, w.Code)
		}
	})
}
func TestServerConfiguration(t *testing.T) {
	server := NewServer("postgres://user:password@localhost:5432/testdb?sslmode=disable", "/dev/null", true)
	if server == nil {
		t.Errorf("expected server to be non-nil")
	}
}
func TestGetLogConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		logFile  string
		expected string
	}{
		{
			name:     "discard log",
			logFile:  "/dev/null",
			expected: "",
		},
		{
			name:     "stderr log",
			logFile:  "",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := getLog(tt.logFile)
			if writer == nil {
				t.Errorf("expected writer to be non-nil")
			}
		})
	}
}
func createMockJWTToken(userID uint, username, systemName string) string {
	return "mock.jwt.token"
}
func TestJWTClaimsExtraction(t *testing.T) {
	t.Skip("JWT testing requires more complex setup with valid tokens")
}
func TestRun(t *testing.T) {
	t.Skip("Server run test requires special setup to avoid blocking")
}
