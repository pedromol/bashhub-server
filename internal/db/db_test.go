package db

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = "postgres://postgres:password@localhost:5432/bashhub_test?sslmode=disable"
	}
	err := setupTestDB(testDBURL)
	if err != nil {
		os.Exit(0)
	}
	createTestTables()
	code := m.Run()
	teardownTestDB()
	os.Exit(code)
}

func setupTestDB(dbURL string) error {
	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	return nil
}

func createTestTables() {
	queries := []string{
		`DROP TABLE IF EXISTS commands CASCADE`,
		`DROP TABLE IF EXISTS systems CASCADE`,
		`DROP TABLE IF EXISTS users CASCADE`,
		`DROP TABLE IF EXISTS configs CASCADE`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(200) UNIQUE NOT NULL,
			email VARCHAR(255),
			password VARCHAR(255),
			registration_code VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS systems (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255),
			mac VARCHAR(255),
			user_id INTEGER REFERENCES users(id),
			hostname VARCHAR(255),
			client_version VARCHAR(255),
			created BIGINT,
			updated BIGINT
		)`,
		`CREATE TABLE IF NOT EXISTS commands (
			id SERIAL PRIMARY KEY,
			command TEXT,
			path VARCHAR(255),
			created BIGINT,
			uuid VARCHAR(255) UNIQUE NOT NULL,
			exit_status INTEGER,
			system_name VARCHAR(255),
			process_id INTEGER,
			process_start_time BIGINT,
			user_id INTEGER REFERENCES users(id),
			session_id VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS configs (
			id SERIAL PRIMARY KEY,
			created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			secret VARCHAR(255)
		)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			panic("Failed to create test table: " + err.Error())
		}
	}
}

func teardownTestDB() {
	if db != nil {
		db.Close()
	}
}

func TestHashAndSalt(t *testing.T) {
	password := "testpassword"
	hashed := HashAndSalt(password)
	if len(hashed) == 0 {
		t.Errorf("expected non-empty value")
	}
	if password == hashed {
		t.Errorf("expected password to be different from hashed version")
	}
	if !strings.Contains(hashed, "$2") {
		t.Errorf("expected hashed password to contain '$2', got %v", hashed)
	}
}

func TestComparePasswords(t *testing.T) {
	password := "testpassword"
	hashed := HashAndSalt(password)
	err := ComparePasswords(hashed, password)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	err = ComparePasswords(hashed, "wrongpassword")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
