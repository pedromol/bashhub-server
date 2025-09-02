package db

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestUser_JSONMarshalUnmarshal(t *testing.T) {
	user := User{
		ID:         1,
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "hashedpassword",
		Mac:        stringPtrModel("AA:BB:CC:DD:EE:FF"),
		SystemName: "test-system",
	}
	jsonData, err := json.Marshal(user)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(jsonData) == 0 {
		t.Errorf("expected non-empty value")
	}
	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, `"id":1`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"id":1`)
	}
	if !strings.Contains(jsonStr, `"Username":"testuser"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"Username":"testuser"`)
	}
	if !strings.Contains(jsonStr, `"email":"test@example.com"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"email":"test@example.com"`)
	}
	if !strings.Contains(jsonStr, `"password":"hashedpassword"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"password":"hashedpassword"`)
	}
	if !strings.Contains(jsonStr, `"mac":"AA:BB:CC:DD:EE:FF"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"mac":"AA:BB:CC:DD:EE:FF"`)
	}
	if !strings.Contains(jsonStr, `"systemName":"test-system"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"systemName":"test-system"`)
	}
	var unmarshaledUser User
	err = json.Unmarshal(jsonData, &unmarshaledUser)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if user.ID != unmarshaledUser.ID {
		t.Errorf("expected %v, got %v", user.ID, unmarshaledUser.ID)
	}
	if user.Username != unmarshaledUser.Username {
		t.Errorf("expected %v, got %v", user.Username, unmarshaledUser.Username)
	}
	if user.Email != unmarshaledUser.Email {
		t.Errorf("expected %v, got %v", user.Email, unmarshaledUser.Email)
	}
	if user.Password != unmarshaledUser.Password {
		t.Errorf("expected %v, got %v", user.Password, unmarshaledUser.Password)
	}
	if *user.Mac != *unmarshaledUser.Mac {
		t.Errorf("expected %v, got %v", *user.Mac, *unmarshaledUser.Mac)
	}
	if user.SystemName != unmarshaledUser.SystemName {
		t.Errorf("expected %v, got %v", user.SystemName, unmarshaledUser.SystemName)
	}
}
func TestUser_WithNilPointers(t *testing.T) {
	user := User{
		ID:         2,
		Username:   "testuser2",
		Email:      "test2@example.com",
		Password:   "hashedpassword2",
		Mac:        nil,
		SystemName: "",
	}
	jsonData, err := json.Marshal(user)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, `"id":2`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"id":2`)
	}
	if !strings.Contains(jsonStr, `"Username":"testuser2"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"Username":"testuser2"`)
	}
	if strings.Contains(jsonStr, `"mac":`) {
		t.Errorf("expected %v to not contain %v", jsonStr, `"mac":`)
	}
	var unmarshaledUser User
	err = json.Unmarshal(jsonData, &unmarshaledUser)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if user.ID != unmarshaledUser.ID {
		t.Errorf("expected %v, got %v", user.ID, unmarshaledUser.ID)
	}
	if user.Username != unmarshaledUser.Username {
		t.Errorf("expected %v, got %v", user.Username, unmarshaledUser.Username)
	}
	if unmarshaledUser.Mac != nil {
		t.Errorf("expected nil, got %v", unmarshaledUser.Mac)
	}
}
func TestQuery_JSONOperations(t *testing.T) {
	query := Query{
		Command:    "ls -la",
		Path:       "/home/user",
		Created:    1640995200,
		Uuid:       "test-uuid-123",
		ExitStatus: 0,
		Username:   "testuser",
		SystemName: "test-system",
		SessionID:  stringPtrModel("session-123"),
	}
	jsonData, err := json.Marshal(query)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, `"command":"ls -la"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"command":"ls -la"`)
	}
	if !strings.Contains(jsonStr, `"path":"/home/user"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"path":"/home/user"`)
	}
	if !strings.Contains(jsonStr, `"created":1640995200`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"created":1640995200`)
	}
	if !strings.Contains(jsonStr, `"uuid":"test-uuid-123"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"uuid":"test-uuid-123"`)
	}
	if !strings.Contains(jsonStr, `"exitStatus":0`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"exitStatus":0`)
	}
	if !strings.Contains(jsonStr, `"username":"testuser"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"username":"testuser"`)
	}
	if !strings.Contains(jsonStr, `"systemName":"test-system"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"systemName":"test-system"`)
	}
	if !strings.Contains(jsonStr, `"sessionId":"session-123"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"sessionId":"session-123"`)
	}
	var unmarshaledQuery Query
	err = json.Unmarshal(jsonData, &unmarshaledQuery)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if query.Command != unmarshaledQuery.Command {
		t.Errorf("expected %v, got %v", query.Command, unmarshaledQuery.Command)
	}
	if query.Path != unmarshaledQuery.Path {
		t.Errorf("expected %v, got %v", query.Path, unmarshaledQuery.Path)
	}
	if query.Created != unmarshaledQuery.Created {
		t.Errorf("expected %v, got %v", query.Created, unmarshaledQuery.Created)
	}
	if query.Uuid != unmarshaledQuery.Uuid {
		t.Errorf("expected %v, got %v", query.Uuid, unmarshaledQuery.Uuid)
	}
	if query.ExitStatus != unmarshaledQuery.ExitStatus {
		t.Errorf("expected %v, got %v", query.ExitStatus, unmarshaledQuery.ExitStatus)
	}
	if query.Username != unmarshaledQuery.Username {
		t.Errorf("expected %v, got %v", query.Username, unmarshaledQuery.Username)
	}
	if query.SystemName != unmarshaledQuery.SystemName {
		t.Errorf("expected %v, got %v", query.SystemName, unmarshaledQuery.SystemName)
	}
	if *query.SessionID != *unmarshaledQuery.SessionID {
		t.Errorf("expected %v, got %v", *query.SessionID, *unmarshaledQuery.SessionID)
	}
}
func TestCommand_JSONOperations(t *testing.T) {
	command := Command{
		ProcessId:        12345,
		ProcessStartTime: 1640995100,
		Uuid:             "cmd-uuid-test",
		Command:          "git status",
		Created:          1640995200,
		Path:             "/home/user/project",
		SystemName:       "laptop",
		ExitStatus:       0,
		User:             User{ID: 1, Username: "testuser"},
		UserId:           1,
		Limit:            50,
		Unique:           true,
		Query:            "git",
		SessionID:        "session-abc",
	}
	jsonData, err := json.Marshal(command)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, `"processId":12345`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"processId":12345`)
	}
	if !strings.Contains(jsonStr, `"processStartTime":1640995100`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"processStartTime":1640995100`)
	}
	if !strings.Contains(jsonStr, `"uuid":"cmd-uuid-test"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"uuid":"cmd-uuid-test"`)
	}
	if !strings.Contains(jsonStr, `"command":"git status"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"command":"git status"`)
	}
	if !strings.Contains(jsonStr, `"created":1640995200`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"created":1640995200`)
	}
	if !strings.Contains(jsonStr, `"path":"/home/user/project"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"path":"/home/user/project"`)
	}
	if !strings.Contains(jsonStr, `"systemName":"laptop"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"systemName":"laptop"`)
	}
	if !strings.Contains(jsonStr, `"exitStatus":0`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"exitStatus":0`)
	}
	if !strings.Contains(jsonStr, `"sessionId":"session-abc"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"sessionId":"session-abc"`)
	}
	var unmarshaledCommand Command
	err = json.Unmarshal(jsonData, &unmarshaledCommand)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if command.ProcessId != unmarshaledCommand.ProcessId {
		t.Errorf("expected %v, got %v", command.ProcessId, unmarshaledCommand.ProcessId)
	}
	if command.Uuid != unmarshaledCommand.Uuid {
		t.Errorf("expected %v, got %v", command.Uuid, unmarshaledCommand.Uuid)
	}
	if command.Command != unmarshaledCommand.Command {
		t.Errorf("expected %v, got %v", command.Command, unmarshaledCommand.Command)
	}
	if command.SessionID != unmarshaledCommand.SessionID {
		t.Errorf("expected %v, got %v", command.SessionID, unmarshaledCommand.SessionID)
	}
}
func TestSystem_JSONOperations(t *testing.T) {
	system := System{
		ID:            1,
		Created:       1640995000,
		Updated:       1640995200,
		Mac:           "AA:BB:CC:DD:EE:FF",
		Hostname:      stringPtrModel("my-laptop"),
		Name:          stringPtrModel("Work Laptop"),
		ClientVersion: stringPtrModel("1.2.3"),
		User:          User{ID: 1},
		UserId:        1,
	}
	jsonData, err := json.Marshal(system)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, `"id":1`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"id":1`)
	}
	if !strings.Contains(jsonStr, `"created":1640995000`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"created":1640995000`)
	}
	if !strings.Contains(jsonStr, `"updated":1640995200`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"updated":1640995200`)
	}
	if !strings.Contains(jsonStr, `"mac":"AA:BB:CC:DD:EE:FF"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"mac":"AA:BB:CC:DD:EE:FF"`)
	}
	if !strings.Contains(jsonStr, `"hostname":"my-laptop"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"hostname":"my-laptop"`)
	}
	if !strings.Contains(jsonStr, `"name":"Work Laptop"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"name":"Work Laptop"`)
	}
	if !strings.Contains(jsonStr, `"clientVersion":"1.2.3"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"clientVersion":"1.2.3"`)
	}
	if !strings.Contains(jsonStr, `"userId":1`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"userId":1`)
	}
	var unmarshaledSystem System
	err = json.Unmarshal(jsonData, &unmarshaledSystem)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if system.ID != unmarshaledSystem.ID {
		t.Errorf("expected %v, got %v", system.ID, unmarshaledSystem.ID)
	}
	if system.Mac != unmarshaledSystem.Mac {
		t.Errorf("expected %v, got %v", system.Mac, unmarshaledSystem.Mac)
	}
	if *system.Hostname != *unmarshaledSystem.Hostname {
		t.Errorf("expected %v, got %v", *system.Hostname, *unmarshaledSystem.Hostname)
	}
	if *system.Name != *unmarshaledSystem.Name {
		t.Errorf("expected %v, got %v", *system.Name, *unmarshaledSystem.Name)
	}
	if *system.ClientVersion != *unmarshaledSystem.ClientVersion {
		t.Errorf("expected %v, got %v", *system.ClientVersion, *unmarshaledSystem.ClientVersion)
	}
}
func TestStatus_JSONOperations(t *testing.T) {
	status := Status{
		User: User{
			ID:       1,
			Username: "testuser",
		},
		ProcessID:            12345,
		Username:             "testuser",
		TotalCommands:        150,
		TotalSessions:        25,
		TotalSystems:         3,
		TotalCommandsToday:   12,
		SessionName:          "bash-session-123",
		SessionStartTime:     1640995000,
		SessionTotalCommands: 8,
	}
	jsonData, err := json.Marshal(status)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, `"username":"testuser"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"username":"testuser"`)
	}
	if !strings.Contains(jsonStr, `"totalCommands":150`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"totalCommands":150`)
	}
	if !strings.Contains(jsonStr, `"totalSessions":25`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"totalSessions":25`)
	}
	if !strings.Contains(jsonStr, `"totalSystems":3`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"totalSystems":3`)
	}
	if !strings.Contains(jsonStr, `"totalCommandsToday":12`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"totalCommandsToday":12`)
	}
	if !strings.Contains(jsonStr, `"sessionName":"bash-session-123"`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"sessionName":"bash-session-123"`)
	}
	if !strings.Contains(jsonStr, `"sessionStartTime":1640995000`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"sessionStartTime":1640995000`)
	}
	if !strings.Contains(jsonStr, `"sessionTotalCommands":8`) {
		t.Errorf("expected %v to contain %v", jsonStr, `"sessionTotalCommands":8`)
	}
	var unmarshaledStatus Status
	err = json.Unmarshal(jsonData, &unmarshaledStatus)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if status.Username != unmarshaledStatus.Username {
		t.Errorf("expected %v, got %v", status.Username, unmarshaledStatus.Username)
	}
	if status.TotalCommands != unmarshaledStatus.TotalCommands {
		t.Errorf("expected %v, got %v", status.TotalCommands, unmarshaledStatus.TotalCommands)
	}
	if status.TotalSessions != unmarshaledStatus.TotalSessions {
		t.Errorf("expected %v, got %v", status.TotalSessions, unmarshaledStatus.TotalSessions)
	}
	if status.SessionName != unmarshaledStatus.SessionName {
		t.Errorf("expected %v, got %v", status.SessionName, unmarshaledStatus.SessionName)
	}
}
func TestImport_TypeAlias(t *testing.T) {
	var imp Import
	var query Query
	if reflect.TypeOf(imp) != reflect.TypeOf(query) {
		t.Errorf("type mismatch")
	}
	if reflect.TypeOf(query) != reflect.TypeOf(imp) {
		t.Errorf("type mismatch")
	}
	importData := Import{
		Command:    "import command",
		Path:       "/import/path",
		Created:    1640995300,
		Uuid:       "import-uuid",
		ExitStatus: 0,
		SystemName: "import-system",
	}
	jsonData, err := json.Marshal(importData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	var unmarshaledImport Import
	err = json.Unmarshal(jsonData, &unmarshaledImport)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if importData.Command != unmarshaledImport.Command {
		t.Errorf("expected %v, got %v", importData.Command, unmarshaledImport.Command)
	}
	if importData.Uuid != unmarshaledImport.Uuid {
		t.Errorf("expected %v, got %v", importData.Uuid, unmarshaledImport.Uuid)
	}
}
func TestModelFieldTags(t *testing.T) {
	t.Run("User model tags", func(t *testing.T) {
		user := User{}
		userType := getFieldTags(user, "Username")
		if !strings.Contains(userType, `json:"Username"`) {
			t.Errorf("expected %v to contain %v", userType, `json:"Username"`)
		}
	})
	t.Run("System model tags", func(t *testing.T) {
		system := System{}
		systemType := getFieldTags(system, "Mac")
		if !strings.Contains(systemType, `gorm:"default:null"`) {
			t.Errorf("expected %v to contain %v", systemType, `gorm:"default:null"`)
		}
		if !strings.Contains(systemType, `json:"mac"`) {
			t.Errorf("expected %v to contain %v", systemType, `json:"mac"`)
		}
	})
}
func TestModelRelationships(t *testing.T) {
	command := Command{
		User: User{ID: 1, Username: "testuser"},
	}
	if uint(1) != command.User.ID {
		t.Errorf("expected %v, got %v", uint(1), command.User.ID)
	}
	if "testuser" != command.User.Username {
		t.Errorf("expected %v, got %v", "testuser", command.User.Username)
	}
	system := System{
		User: User{ID: 2, Username: "systemuser"},
	}
	if uint(2) != system.User.ID {
		t.Errorf("expected %v, got %v", uint(2), system.User.ID)
	}
	if "systemuser" != system.User.Username {
		t.Errorf("expected %v, got %v", "systemuser", system.User.Username)
	}
}
func TestModelDefaultValues(t *testing.T) {
	command := Command{}
	if 0 != command.ExitStatus {
		t.Errorf("expected %v, got %v", 0, command.ExitStatus)
	}
	if 0 != command.Limit {
		t.Errorf("expected %v, got %v", 0, command.Limit)
	}
	if command.Unique {
		t.Errorf("expected false, got true")
	}
	if len(command.Query) != 0 {
		t.Errorf("expected empty, got %v", command.Query)
	}
	system := System{}
	if len(system.Mac) != 0 {
		t.Errorf("expected empty, got %v", system.Mac)
	}
	if system.Hostname != nil {
		t.Errorf("expected nil, got %v", system.Hostname)
	}
	if system.Name != nil {
		t.Errorf("expected nil, got %v", system.Name)
	}
	if system.ClientVersion != nil {
		t.Errorf("expected nil, got %v", system.ClientVersion)
	}
}
func TestJSONEdgeCases(t *testing.T) {
	t.Run("empty strings", func(t *testing.T) {
		user := User{Username: "", Email: "", Password: ""}
		jsonData, err := json.Marshal(user)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		var unmarshaled User
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(unmarshaled.Username) != 0 {
			t.Errorf("expected empty, got %v", unmarshaled.Username)
		}
		if len(unmarshaled.Email) != 0 {
			t.Errorf("expected empty, got %v", unmarshaled.Email)
		}
		if len(unmarshaled.Password) != 0 {
			t.Errorf("expected empty, got %v", unmarshaled.Password)
		}
	})
	t.Run("zero values", func(t *testing.T) {
		query := Query{
			Created:    0,
			ExitStatus: 0,
		}
		jsonData, err := json.Marshal(query)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		jsonStr := string(jsonData)
		if !strings.Contains(jsonStr, `"created":0`) {
			t.Errorf("expected %v to contain %v", jsonStr, `"created":0`)
		}
		if !strings.Contains(jsonStr, `"exitStatus":0`) {
			t.Errorf("expected %v to contain %v", jsonStr, `"exitStatus":0`)
		}
	})
}
func TestModelValidation(t *testing.T) {
	t.Run("User validation", func(t *testing.T) {
		validUser := User{
			Username: "validuser",
			Email:    "valid@example.com",
			Password: "validpass",
		}
		if len(validUser.Username) == 0 {
			t.Errorf("expected non-empty value")
		}
		if len(validUser.Email) == 0 {
			t.Errorf("expected non-empty value")
		}
		if len(validUser.Password) == 0 {
			t.Errorf("expected non-empty value")
		}
		invalidUser := User{}
		if len(invalidUser.Username) != 0 {
			t.Errorf("expected empty, got %v", invalidUser.Username)
		}
		if len(invalidUser.Email) != 0 {
			t.Errorf("expected empty, got %v", invalidUser.Email)
		}
		if len(invalidUser.Password) != 0 {
			t.Errorf("expected empty, got %v", invalidUser.Password)
		}
	})
	t.Run("Command validation", func(t *testing.T) {
		validCommand := Command{
			Uuid:    "valid-uuid",
			Command: "valid command",
			User:    User{ID: 1},
		}
		if len(validCommand.Uuid) == 0 {
			t.Errorf("expected non-empty value")
		}
		if len(validCommand.Command) == 0 {
			t.Errorf("expected non-empty value")
		}
		if validCommand.User.ID == 0 {
			t.Errorf("expected non-zero value")
		}
	})
}
func getFieldTags(model interface{}, fieldName string) string {
	return ""
}
func stringPtrModel(s string) *string {
	return &s
}
