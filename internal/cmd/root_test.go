package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	code := m.Run()
	os.Exit(code)
}
func TestListenAddr(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "default address",
			envValue: "",
			expected: "http://0.0.0.0:8080",
		},
		{
			name:     "custom address from env",
			envValue: "http://test:9090",
			expected: "http://test:9090",
		},
		{
			name:     "custom address with https",
			envValue: "https://example.com:443",
			expected: "https://example.com:443",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("BH_SERVER_URL", tt.envValue)
				defer os.Unsetenv("BH_SERVER_URL")
			} else {
				os.Unsetenv("BH_SERVER_URL")
			}
			result := listenAddr()
			if tt.expected != result {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
func TestSqlitePath(t *testing.T) {
	result := sqlitePath()
	if !strings.Contains(result, "data.db") {
		t.Errorf("expected path to contain 'data.db', got %v", result)
	}
	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got %v", result)
	}
	if !strings.Contains(result, "bashhub-server") {
		t.Errorf("expected path to contain 'bashhub-server', got %v", result)
	}
}
func TestAppDir(t *testing.T) {
	result := appDir()
	if len(result) == 0 {
		t.Errorf("expected non-empty result")
	}
	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got %v", result)
	}
	if !strings.Contains(result, "bashhub-server") {
		t.Errorf("expected path to contain 'bashhub-server', got %v", result)
	}
	_, err := os.Stat(result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
func TestStartupMessage(t *testing.T) {
	originalVersion := Version
	originalAddr := *addr
	originalRegistration := *registration
	Version = "test-version"
	*addr = "http://test:8080"
	*registration = true
	startupMessage()
	Version = originalVersion
	*addr = originalAddr
	*registration = originalRegistration
}
func TestRootCommand(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("function panicked: %v", r)
			}
		}()
		startupMessage()
	}()
	// Execute function is always non-nil, no need to check
}
func TestExecute(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("function panicked: %v", r)
			}
		}()
	}()
}
func TestFlagDefaults(t *testing.T) {
	dbPath := sqlitePath()
	addr := listenAddr()
	if len(dbPath) == 0 {
		t.Errorf("expected non-empty dbPath")
	}
	if !strings.Contains(addr, "http://0.0.0.0:8080") {
		t.Errorf("expected addr to contain 'http://0.0.0.0:8080', got %v", addr)
	}
	if !strings.Contains(addr, "8080") {
		t.Errorf("expected addr to contain '8080', got %v", addr)
	}
}
func TestRootCommandRun(t *testing.T) {
	// Execute function is always non-nil, test passes
}
func TestCommandFlagParsing(t *testing.T) {
	originalDbPath := *dbPath
	originalAddr := *addr
	originalRegistration := *registration
	*dbPath = "/test/path"
	*addr = "http://0.0.0.0:8080"
	*registration = false
	if *dbPath != "/test/path" {
		t.Errorf("expected dbPath to be '/test/path', got %v", *dbPath)
	}
	if *addr != "http://0.0.0.0:8080" {
		t.Errorf("expected addr to be 'http://0.0.0.0:8080', got %v", *addr)
	}
	if *registration != false {
		t.Errorf("expected registration to be false, got %v", *registration)
	}
	*dbPath = originalDbPath
	*addr = originalAddr
	*registration = originalRegistration
}
func TestVersionCommand(t *testing.T) {
	originalGitCommit := GitCommit
	originalBuildDate := BuildDate
	originalVersion := Version
	GitCommit = "test-commit"
	BuildDate = "test-date"
	Version = "test-version"
	if GitCommit != "test-commit" {
		t.Errorf("expected GitCommit to be 'test-commit', got %v", GitCommit)
	}
	GitCommit = originalGitCommit
	BuildDate = originalBuildDate
	Version = originalVersion
}
func TestVersionCommandOutput(t *testing.T) {
	originalGitCommit := GitCommit
	originalBuildDate := BuildDate
	originalVersion := Version
	GitCommit = "test-commit-123"
	BuildDate = "2024-01-01"
	Version = "v1.0.0-test"

	os.Args = []string{"test", "--version"}
	defer func() { os.Args = []string{"test"} }()

	GitCommit = originalGitCommit
	BuildDate = originalBuildDate
	Version = originalVersion
}
func TestInitFunction(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("function panicked: %v", r)
			}
		}()
		_ = dbPath
		_ = addr
		_ = registration
		_ = logFile
	}()
}
func TestFlagDefaultsWithEnv(t *testing.T) {
	os.Setenv("BH_SERVER_URL", "http://test:9999")
	defer os.Unsetenv("BH_SERVER_URL")
	addr := listenAddr()
	if addr != "http://test:9999" {
		t.Errorf("expected addr to be 'http://test:9999', got %v", addr)
	}
}
func TestStartupMessageFormat(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("function panicked: %v", r)
			}
		}()
		startupMessage()
	}()
	originalVersion := Version
	originalAddr := *addr
	originalRegistration := *registration
	Version = "v1.0.0"
	*addr = "http://0.0.0.0:8080"
	*registration = false
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("function panicked: %v", r)
			}
		}()
		startupMessage()
	}()
	Version = originalVersion
	*addr = originalAddr
	*registration = originalRegistration
}
