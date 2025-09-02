package server

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pedromol/bashhub-server/internal/db"
)

func getLog(logFile string) io.Writer {
	switch {
	case logFile == "/dev/null":
		return io.Discard
	case logFile != "":
		f, err := os.Create(logFile)
		if err != nil {
			log.Fatal(err)
		}
		return f
	default:
		return os.Stderr
	}
}

type Server struct {
	mux          *http.ServeMux
	dbPath       string
	logFile      string
	registration bool
	logger       io.Writer
	jwtSecret    []byte
}

func NewServer(dbPath, logFile string, registration bool) *Server {
	secret := generateSecret()
	s := &Server{
		mux:          http.NewServeMux(),
		dbPath:       dbPath,
		logFile:      logFile,
		registration: registration,
		logger:       getLog(logFile),
		jwtSecret:    []byte(secret),
	}

	s.setupRoutes()
	return s
}

func generateSecret() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.StdEncoding.EncodeToString(bytes)
}

func (s *Server) setupRoutes() {
	db.Init(s.dbPath)

	s.mux.HandleFunc("/ping", s.handlePing)
	s.mux.HandleFunc("/api/v1/login", s.handleLogin)
	s.mux.HandleFunc("/api/v1/user", s.handleUserCreate)
	s.mux.HandleFunc("/api/v1/command/", s.authMiddleware(s.handleCommand))
	s.mux.HandleFunc("/api/v1/system", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.authMiddleware(s.handleSystemCreate)(w, r)
		} else if r.Method == http.MethodGet {
			s.authMiddleware(s.handleSystemGet)(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	s.mux.HandleFunc("/api/v1/system/", s.authMiddleware(s.handleSystemUpdate))
	s.mux.HandleFunc("/api/v1/client-view/status", s.authMiddleware(s.handleStatus))
	s.mux.HandleFunc("/api/v1/import", s.authMiddleware(s.handleImport))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	s.mux.ServeHTTP(w, r)
	s.logRequest(r, time.Since(start))
}

func (s *Server) logRequest(r *http.Request, duration time.Duration) {
	fmt.Fprintf(s.logger, "[BASHHUB-SERVER] %s | %s | %v | %s | %s\n",
		time.Now().Format("2006/01/02 - 15:04:05"),
		r.Method,
		duration,
		r.RemoteAddr,
		r.URL.Path,
	)
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user db.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		s.respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := user.UserExists(); err != nil {
		s.respondError(w, r, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	user.SystemName, _ = user.UserGetSystemName()
	user.ID, _ = user.UserGetID()

	token, err := s.generateToken(&user)
	if err != nil {
		s.respondError(w, r, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"accessToken": token})
}

func (s *Server) handleUserCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.registration {
		s.respondError(w, r, http.StatusForbidden, "Registration of new users is not allowed")
		return
	}

	var user db.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		s.respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if user.Email == "" {
		s.respondError(w, r, http.StatusBadRequest, "email required")
		return
	}

	if exists, _ := user.UsernameExists(); exists {
		s.respondError(w, r, http.StatusConflict, "Username already taken")
		return
	}

	if exists, _ := user.EmailExists(); exists {
		s.respondError(w, r, http.StatusConflict, "This email address is already registered")
		return
	}

	_, err := user.UserCreate()
	if err != nil {
		s.respondError(w, r, http.StatusInternalServerError, "Failed to create user")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.respondError(w, r, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			s.respondError(w, r, http.StatusUnauthorized, "Invalid authorization header")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		user, err := s.validateToken(token)
		if err != nil {
			s.respondError(w, r, http.StatusUnauthorized, "Invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) respondError(w http.ResponseWriter, r *http.Request, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    status,
		"message": message,
	})
}

func (s *Server) generateToken(user *db.User) (string, error) {
	header := base64.StdEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := fmt.Sprintf(`{"username":"%s","systemName":"%s","user_id":%d,"exp":%d}`,
		user.Username, user.SystemName, user.ID, time.Now().Add(10000*time.Hour).Unix())

	headerPayload := header + "." + base64.StdEncoding.EncodeToString([]byte(payload))

	mac := hmac.New(sha256.New, s.jwtSecret)
	mac.Write([]byte(headerPayload))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return headerPayload + "." + signature, nil
}

func (s *Server) validateToken(tokenString string) (*db.User, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerPayload := parts[0] + "." + parts[1]
	expectedSignature, _ := base64.StdEncoding.DecodeString(parts[2])

	mac := hmac.New(sha256.New, s.jwtSecret)
	mac.Write([]byte(headerPayload))
	actualSignature := mac.Sum(nil)

	if !hmac.Equal(expectedSignature, actualSignature) {
		return nil, fmt.Errorf("invalid token signature")
	}

	payloadBytes, _ := base64.StdEncoding.DecodeString(parts[1])
	var claims struct {
		Username   string `json:"username"`
		SystemName string `json:"systemName"`
		UserID     uint   `json:"user_id"`
		Exp        int64  `json:"exp"`
	}

	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("invalid token payload")
	}

	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &db.User{
		Username:   claims.Username,
		SystemName: claims.SystemName,
		ID:         claims.UserID,
	}, nil
}

func (s *Server) handleCommand(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*db.User)

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/command/")
	if path == "search" {
		s.handleCommandSearch(w, r, user)
	} else {
		s.handleCommandGet(w, r, user, path)
	}
}

func (s *Server) handleCommandSearch(w http.ResponseWriter, r *http.Request, user *db.User) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cmd := db.Command{
		User:     *user,
		Limit:    100,
		Unique:   r.URL.Query().Get("unique") == "true",
		Query:    r.URL.Query().Get("query"),
		Path:     r.URL.Query().Get("path"),
		SystemName: r.URL.Query().Get("systemName"),
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			cmd.Limit = l
		}
	}

	results, err := cmd.CommandGet()
	if err != nil {
		s.respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(results) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{})
	} else {
		json.NewEncoder(w).Encode(results)
	}
}

func (s *Server) handleCommandGet(w http.ResponseWriter, r *http.Request, user *db.User, uuid string) {
	if r.Method != http.MethodGet && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cmd := db.Command{
		User: *user,
		Uuid: uuid,
	}

	if r.Method == http.MethodDelete {
		_, err := cmd.CommandDelete()
		if err != nil {
			s.respondError(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	result, err := cmd.CommandGetUUID()
	if err != nil {
		s.respondError(w, r, http.StatusNotFound, "Command not found")
		return
	}

	result.Username = user.Username
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleSystemCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := r.Context().Value("user").(*db.User)

	var system db.System
	if err := json.NewDecoder(r.Body).Decode(&system); err != nil {
		s.respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	system.User = *user
	system.Created = time.Now().Unix()
	system.Updated = time.Now().Unix()

	_, err := system.SystemInsert()
	if err != nil {
		s.respondError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleSystemGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := r.Context().Value("user").(*db.User)
	mac := r.URL.Query().Get("mac")
	if mac == "" {
		s.respondError(w, r, http.StatusBadRequest, "mac parameter required")
		return
	}

	system := db.System{
		User: *user,
		Mac:  mac,
	}

	result, err := system.SystemGet()
	if err != nil {
		s.respondError(w, r, http.StatusNotFound, "System not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleSystemUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := r.Context().Value("user").(*db.User)
	mac := strings.TrimPrefix(r.URL.Path, "/api/v1/system/")

	var system db.System
	if err := json.NewDecoder(r.Body).Decode(&system); err != nil {
		s.respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	system.User = *user
	system.Mac = mac
	system.Updated = time.Now().Unix()

	_, err := system.SystemUpdate()
	if err != nil {
		s.respondError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := r.Context().Value("user").(*db.User)

	status := db.Status{
		User:      *user,
		ProcessID: 0,
	}

	if pid := r.URL.Query().Get("processId"); pid != "" {
		if p, err := strconv.Atoi(pid); err == nil {
			status.ProcessID = p
		}
	}

	result, err := status.StatusGet()
	if err != nil {
		s.respondError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := r.Context().Value("user").(*db.User)

	var imp db.Import
	if err := json.NewDecoder(r.Body).Decode(&imp); err != nil {
		s.respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	imp.Username = user.Username
	err := db.ImportCommands(imp)
	if err != nil {
		s.respondError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func Run(dbFile string, logFile string, addr string, registration bool) {
	server := NewServer(dbFile, logFile, registration)
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, server))
}
