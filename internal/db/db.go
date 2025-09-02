package db

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

const (
	errCheckingRowExists = "error checking if row exists %v"
)

var (
	db              *sql.DB
	connectionLimit int
)

func Init(dbPath string) error {
	var err error
	db, err = sql.Open("postgres", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	connectionLimit = 50
	db.SetMaxOpenConns(connectionLimit)
	createTables()
	createIndexes()
	return nil
}

func createTables() {
	queries := []string{
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
			log.Fatalf("Failed to create table: %v", err)
		}
	}
}

func createIndexes() {
	indexes := []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_user ON users(username)",
		"CREATE INDEX IF NOT EXISTS idx_mac ON systems(mac)",
		"CREATE INDEX IF NOT EXISTS idx_user_command_created ON commands(user_id, created, command)",
		"CREATE INDEX IF NOT EXISTS idx_user_uuid ON commands(user_id, uuid)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_uuid ON commands(uuid)",
	}

	for _, index := range indexes {
		_, err := db.Exec(index)
		if err != nil {
			log.Fatalf("Failed to create index: %v", err)
		}
	}
}
func GetSecret() (string, error) {
	var err error
	var secret string
	if connectionLimit != 1 {
		_, err = db.Exec(`INSERT INTO configs ("id","created", "secret") 
						VALUES (1, now(), (SELECT md5(random()::text)))
						ON conflict do nothing;`)
	} else {
		_, err = db.Exec(`INSERT INTO configs ("id","created" ,"secret") 
						VALUES (1, current_timestamp, lower(hex(randomblob(16)))) 
						ON conflict do nothing;`)
	}
	if err != nil {
		log.Fatal(err)
	}
	err = db.QueryRow(`SELECT "secret" from configs where "id" = 1 `).Scan(&secret)
	if err != nil {
		log.Fatal(err)
	}
	return secret, nil
}
func HashAndSalt(password string) string {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		log.Println("Failed to generate salt:", err)
		return ""
	}

	hash := sha256.Sum256(append([]byte(password), salt...))
	result := append(salt, hash[:]...)
	return hex.EncodeToString(result)
}

func ComparePasswords(hashedPwd string, plainPwd string) error {
	data, err := hex.DecodeString(hashedPwd)
	if err != nil {
		return err
	}

	if len(data) < 32 {
		return fmt.Errorf("invalid hash format")
	}

	salt := data[:16]
	storedHash := data[16:]

	hash := sha256.Sum256(append([]byte(plainPwd), salt...))

	if !equalHashes(hash[:], storedHash) {
		return fmt.Errorf("password mismatch")
	}

	return nil
}

func equalHashes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
func (user User) UserExists() error {
	var password string
	err := db.QueryRow("SELECT password FROM users WHERE username = $1",
		user.Username).Scan(&password)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf(errCheckingRowExists, err)
	}
	if password != "" {
		return ComparePasswords(password, user.Password)
	}
	return nil
}
func (user User) UserGetID() (uint, error) {
	var id uint
	err := db.QueryRow(`SELECT "id" 
							FROM users 
							WHERE "username"  = $1`,
		user.Username).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf(errCheckingRowExists, err)
	}
	return id, nil
}
func (user User) UserGetSystemName() (string, error) {
	var systemName string
	err := db.QueryRow(`SELECT name 
							FROM systems 
							WHERE user_id in (select id from users where username = $1)
							AND mac = $2`,
		user.Username, user.Mac).Scan(&systemName)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf(errCheckingRowExists, err)
	}
	return systemName, nil
}
func (user User) UsernameExists() (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT exists (select id FROM users WHERE "username" = $1)`,
		user.Username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf(errCheckingRowExists, err)
	}
	return exists, nil
}
func (user User) EmailExists() (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT exists (select id FROM users WHERE "email" = $1)`,
		user.Email).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf(errCheckingRowExists, err)
	}
	return exists, nil
}
func (user User) UserCreate() (int64, error) {
	user.Password = HashAndSalt(user.Password)
	res, err := db.Exec(`INSERT INTO users("registration_code", "username","password","email")
 							 VALUES ($1,$2,$3,$4) ON CONFLICT(username) do nothing`, user.RegistrationCode,
		user.Username, user.Password, user.Email)
	if err != nil {
		log.Fatal(err)
	}
	inserted, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	return inserted, nil
}
func (cmd Command) CommandInsert() (int64, error) {
	res, err := db.Exec(`
	INSERT INTO commands("process_id","process_start_time","exit_status","uuid","command", "created", "path", "user_id", "system_name")
 	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT do nothing`,
		cmd.ProcessId, cmd.ProcessStartTime, cmd.ExitStatus, cmd.Uuid, cmd.Command, cmd.Created, cmd.Path, cmd.User.ID, cmd.SystemName)
	if err != nil {
		log.Fatal(err)
	}
	inserted, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	return inserted, nil
}
func (cmd Command) CommandGet() ([]Query, error) {
	var (
		results []Query
	)
	var query string
	var args []interface{}
	if cmd.Query != "" {
		if cmd.Unique {
			if cmd.Path != "" && cmd.SystemName != "" {
				query = `
					SELECT DISTINCT ON ("command") command, "uuid", "created"
					FROM commands
					WHERE "user_id" = $1 AND "path" = $2 AND "system_name" = $3 AND "command" ~ $4
					ORDER BY "command", "created" DESC LIMIT $5`
				args = []interface{}{cmd.User.ID, cmd.Path, cmd.SystemName, cmd.Query, cmd.Limit}
			} else if cmd.Path != "" {
				query = `
					SELECT DISTINCT ON ("command") command, "uuid", "created"
					FROM commands
					WHERE "user_id" = $1 AND "path" = $2 AND "command" ~ $3
					ORDER BY "command", "created" DESC LIMIT $4`
				args = []interface{}{cmd.User.ID, cmd.Path, cmd.Query, cmd.Limit}
			} else if cmd.SystemName != "" {
				query = `
					SELECT DISTINCT ON ("command") command, "uuid", "created"
					FROM commands
					WHERE "user_id" = $1 AND "system_name" = $2 AND "command" ~ $3
					ORDER BY "command", "created" DESC LIMIT $4`
				args = []interface{}{cmd.User.ID, cmd.SystemName, cmd.Query, cmd.Limit}
			} else {
				query = `
					SELECT DISTINCT ON ("command") command, "uuid", "created"
					FROM commands
					WHERE "user_id" = $1 AND "command" ~ $2
					ORDER BY "command", "created" DESC LIMIT $3`
				args = []interface{}{cmd.User.ID, cmd.Query, cmd.Limit}
			}
		} else {
			if cmd.Path != "" && cmd.SystemName != "" {
				query = `
					SELECT "command", "uuid", "created"
					FROM commands
					WHERE "user_id" = $1 AND "path" = $2 AND "system_name" = $3 AND "command" ~ $4
					ORDER BY "created" DESC LIMIT $5`
				args = []interface{}{cmd.User.ID, cmd.Path, cmd.SystemName, cmd.Query, cmd.Limit}
			} else if cmd.Path != "" {
				query = `
					SELECT "command", "uuid", "created"
					FROM commands
					WHERE "user_id" = $1 AND "path" = $2 AND "command" ~ $3
					ORDER BY "created" DESC LIMIT $4`
				args = []interface{}{cmd.User.ID, cmd.Path, cmd.Query, cmd.Limit}
			} else if cmd.SystemName != "" {
				query = `
					SELECT "command", "uuid", "created"
					FROM commands
					WHERE "user_id" = $1 AND "system_name" = $2 AND "command" ~ $3
					ORDER BY "created" DESC LIMIT $4`
				args = []interface{}{cmd.User.ID, cmd.SystemName, cmd.Query, cmd.Limit}
			} else {
				query = `
					SELECT "command", "uuid", "created"
					FROM commands
					WHERE "user_id" = $1 AND "command" ~ $2
					ORDER BY "created" DESC LIMIT $3`
				args = []interface{}{cmd.User.ID, cmd.Query, cmd.Limit}
			}
		}
	} else if cmd.Unique {
		if cmd.Path != "" && cmd.SystemName != "" {
			query = `
				SELECT DISTINCT ON ("command") command, "uuid", "created"
				FROM commands
				WHERE "user_id" = $1 AND "path" = $2 AND "system_name" = $3
				ORDER BY "command", "created" DESC LIMIT $4`
			args = []interface{}{cmd.User.ID, cmd.Path, cmd.SystemName, cmd.Limit}
		} else if cmd.Path != "" {
			query = `
				SELECT DISTINCT ON ("command") command, "uuid", "created"
				FROM commands
				WHERE "user_id" = $1 AND "path" = $2
				ORDER BY "command", "created" DESC LIMIT $3`
			args = []interface{}{cmd.User.ID, cmd.Path, cmd.Limit}
		} else if cmd.SystemName != "" {
			query = `
				SELECT DISTINCT ON ("command") command, "uuid", "created"
				FROM commands
				WHERE "user_id" = $1 AND "system_name" = $2
				ORDER BY "command", "created" DESC LIMIT $3`
			args = []interface{}{cmd.User.ID, cmd.SystemName, cmd.Limit}
		} else {
			query = `
				SELECT DISTINCT ON ("command") command, "uuid", "created"
				FROM commands
				WHERE "user_id" = $1
				ORDER BY "command", "created" DESC LIMIT $2`
			args = []interface{}{cmd.User.ID, cmd.Limit}
		}
	} else {
		if cmd.Path != "" && cmd.SystemName != "" {
			query = `
				SELECT "command", "uuid", "created"
				FROM commands
				WHERE "user_id" = $1 AND "path" = $2 AND "system_name" = $3
				ORDER BY "created" DESC LIMIT $4`
			args = []interface{}{cmd.User.ID, cmd.Path, cmd.SystemName, cmd.Limit}
		} else if cmd.Path != "" {
			query = `
				SELECT "command", "uuid", "created"
				FROM commands
				WHERE "user_id" = $1 AND "path" = $2
				ORDER BY "created" DESC LIMIT $3`
			args = []interface{}{cmd.User.ID, cmd.Path, cmd.Limit}
		} else if cmd.SystemName != "" {
			query = `
				SELECT "command", "uuid", "created"
				FROM commands
				WHERE "user_id" = $1 AND "system_name" = $2
				ORDER BY "created" DESC LIMIT $3`
			args = []interface{}{cmd.User.ID, cmd.SystemName, cmd.Limit}
		} else {
			query = `
				SELECT "command", "uuid", "created"
				FROM commands
				WHERE "user_id" = $1
				ORDER BY "created" DESC LIMIT $2`
			args = []interface{}{cmd.User.ID, cmd.Limit}
		}
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return []Query{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var result Query
		err = rows.Scan(&result.Command, &result.Uuid, &result.Created)
		if err != nil {
			return []Query{}, err
		}
		results = append(results, result)
	}
	return results, nil
}
func (cmd Command) CommandGetUUID() (Query, error) {
	var result Query
	err := db.QueryRow(`
	SELECT "command","path", "created" , "uuid", "exit_status", "system_name", "process_id" 
		FROM commands
		WHERE "uuid" = $1 
	AND "user_id" = $2`, cmd.Uuid, cmd.User.ID).Scan(&result.Command, &result.Path, &result.Created, &result.Uuid,
		&result.ExitStatus, &result.SystemName, &result.SessionID)
	if err != nil {
		return Query{}, err
	}
	return result, nil
}
func (cmd Command) CommandDelete() (int64, error) {
	res, err := db.Exec(`
	DELETE FROM commands WHERE "user_id" = $1 AND "uuid" = $2 `, cmd.User.ID, cmd.Uuid)
	if err != nil {
		log.Fatal(err)
	}
	inserted, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	return inserted, nil
}
func (sys System) SystemUpdate() (int64, error) {
	t := time.Now().Unix()
	res, err := db.Exec(`
	UPDATE systems 
		SET "hostname" = $1 , "updated" = $2
		WHERE "user_id" = $3
		AND "mac" = $4`,
		sys.Hostname, t, sys.User.ID, sys.Mac)
	if err != nil {
		log.Fatal(err)
	}
	inserted, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	return inserted, nil
}
func (sys System) SystemInsert() (int64, error) {
	t := time.Now().Unix()
	res, err := db.Exec(`INSERT INTO systems ("name", "mac", "user_id", "hostname", "client_version", "created", "updated")
 									  VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		sys.Name, sys.Mac, sys.User.ID, sys.Hostname, sys.ClientVersion, t, t)
	if err != nil {
		log.Fatal(err)
	}
	inserted, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	return inserted, nil
}
func (sys System) SystemGet() (System, error) {
	var row System
	err := db.QueryRow(`SELECT "name", "mac", "user_id", "hostname", "client_version",
 									  "id", "created", "updated" FROM systems 
 							  WHERE  "user_id" = $1
 							  AND "mac" = $2`,
		sys.User.ID, sys.Mac).Scan(&row.Name, &row.Mac, &row.UserId, &row.Hostname,
		&row.ClientVersion, &row.ID, &row.Created, &row.Updated)
	if err != nil {
		return System{}, err
	}
	return row, nil
}
func (status Status) StatusGet() (Status, error) {
	err := db.QueryRow(`
		select
      		( select count(*) from commands where user_id = $1) as totalCommands,
      		( select count(distinct process_id) from commands where user_id = $1) as totalSessions,
      		( select count(*) from systems where user_id = $1) as totalSystems,
      		( select count(*) from commands where to_timestamp(cast(created/1000 as bigint))::date = now()::date and  user_id = $1) as totalCommandsToday,
      		( select count(*) from commands where process_id = $2) as sessionTotalCommands`,
		status.User.ID, status.ProcessID).Scan(
		&status.TotalCommands, &status.TotalSessions, &status.TotalSystems,
		&status.TotalCommandsToday, &status.SessionTotalCommands)
	if err != nil {
		return Status{}, err
	}
	return status, err
}
func ImportCommands(imp Import) error {
	_, err := db.Exec(`
	INSERT INTO commands ("command", "path", "created", "uuid", "exit_status","system_name", "session_id", "user_id" )
	VALUES ($1,$2,$3,$4,$5,$6,$7 ,(select "id" from users where "username" = $8)) ON CONFLICT do nothing`,
		imp.Command, imp.Path, imp.Created, imp.Uuid, imp.ExitStatus, imp.SystemName, imp.SessionID, imp.Username)
	if err != nil {
		return err
	}
	return nil
}
