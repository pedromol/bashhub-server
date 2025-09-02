# Security Measures - SQL Injection Prevention

## Overview

This codebase implements comprehensive protection against SQL injection attacks. All database operations use parameterized queries to ensure user input cannot manipulate SQL statements.

## Security Features

### 1. Parameterized Queries
All database queries use PostgreSQL's parameterized query syntax with `$1`, `$2`, `$3` placeholders:

```go
// ✅ SAFE: Parameterized query
rows, err := db.Query("SELECT * FROM users WHERE username = $1", username)

// ❌ UNSAFE: String concatenation (vulnerable to SQL injection)
query := "SELECT * FROM users WHERE username = '" + username + "'"
```

### 2. Automatic Parameter Escaping
The PostgreSQL driver (`pq`) automatically escapes all parameters, preventing:
- Single quote escapes
- Double quote escapes
- Unicode character attacks
- Special character injections

### 3. Comprehensive Test Coverage
The test suite includes extensive SQL injection prevention tests covering:

#### Attack Vectors Tested
- **Classic SQL Injection**: `'; DROP TABLE users; --`
- **Tautology Attacks**: `' OR '1'='1`
- **Union-Based Attacks**: `' UNION SELECT * FROM users; --`
- **Comment Attacks**: `admin' --`
- **Stacked Queries**: `'; DELETE FROM commands; --`

#### Advanced Attack Vectors
- **Unicode Injection**: Unicode quotes and special characters
- **Encoding Attacks**: URL-encoded, hex-encoded, base64 payloads
- **Time-Based Attacks**: `pg_sleep()` and timing functions
- **Error-Based Attacks**: Functions causing database errors
- **Second-Order Injection**: Stored and nested injection patterns

#### Database-Specific Attacks
- **PostgreSQL Functions**: `pg_sleep()`, `pg_shadow`, version queries
- **Stored Procedures**: `EXEC` statements and procedure calls
- **PostgreSQL Operators**: Database-specific syntax attacks

## Protected Functions

All database functions that accept user input are protected:

- `UserExists()` - Username validation
- `UserGetID()` - User ID retrieval
- `UsernameExists()` - Username existence check
- `EmailExists()` - Email validation
- `CommandGet()` - Command search with regex
- `CommandGetUUID()` - UUID-based command retrieval
- `CommandDelete()` - Command deletion
- `UserCreate()` - User registration
- `CommandInsert()` - Command storage

## Testing

### Running Security Tests

```bash
# Run all SQL injection prevention tests
make test-sql-injection

# Run all security-related tests
make test-security

# Run specific SQL injection test
go test -v ./internal/db/... -run "TestSQLInjectionPrevention_CommandGet"
```

### Test Coverage

The security tests cover:
- **38 different SQL injection attack vectors**
- **All database functions with user input**
- **Parameter validation and escaping**
- **Regression prevention tests**

## Development Guidelines

### ✅ DO Use
```go
// Safe parameterized queries
db.Query("SELECT * FROM table WHERE field = $1 AND other = $2", val1, val2)
db.Exec("INSERT INTO table VALUES ($1, $2, $3)", val1, val2, val3)
```

### ❌ DON'T Use
```go
// Dangerous string concatenation
query := "SELECT * FROM table WHERE field = '" + input + "'"

// Dangerous fmt.Sprintf
query := fmt.Sprintf("SELECT * FROM table WHERE field = '%s'", input)

// Dangerous string replacement
query := strings.Replace("SELECT * FROM table WHERE field = 'PLACEHOLDER'", "PLACEHOLDER", input, 1)
```

## Security Assurance

### Automated Testing
- All SQL injection tests run automatically with `make test`
- Tests fail if SQL injection vulnerabilities are reintroduced
- Comprehensive coverage of attack vectors

### Code Review Checks
- Static analysis prevents unsafe SQL patterns
- Parameterized query validation
- Input sanitization verification

### Continuous Security
- Security tests are part of CI/CD pipeline
- Regression tests prevent vulnerability reintroduction
- Documentation for secure coding practices

## Incident Response

If a SQL injection vulnerability is discovered:

1. **Immediate Action**: Add test case that exposes the vulnerability
2. **Fix Implementation**: Convert to parameterized query
3. **Verification**: Ensure new test passes and existing tests still work
4. **Documentation**: Update this security guide if needed

## References

- [OWASP SQL Injection Prevention Cheat Sheet](https://owasp.org/www-community/attacks/SQL_Injection)
- [PostgreSQL Parameterized Queries](https://www.postgresql.org/docs/current/sql-prepare.html)
- [Go database/sql Package](https://golang.org/pkg/database/sql/)
