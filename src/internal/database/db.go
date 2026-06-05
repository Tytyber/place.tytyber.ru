package database

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"place-tytyber/internal/models"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

// Session state for interactive commands
var sessionState *models.SessionState

func InitDB(host string, port int, user string, password string, dbname string) {
	// Формируем URL правильно, даже если пароль пустой
	var connStr string
	if password == "" {
		connStr = fmt.Sprintf(
			"postgres://%s@%s:%d/%s?sslmode=disable",
			user, host, port, dbname,
		)
	} else {
		connStr = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			user, password, host, port, dbname,
		)
	}

	log.Printf("🔍 DEBUG connStr: postgres://%s:***@%s:%d/%s", user, host, port, dbname)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Printf("Successfully connected to database: %s", dbname)
	createTables()
}

func createTables() {
	// Check if users table exists
	var tableName string
	err := db.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_name = 'users'").Scan(&tableName)

	if err == sql.ErrNoRows {
		// Table doesn't exist, create it
		createUsersTable := `
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255),
			password_hash TEXT DEFAULT '',
			is_guest BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`

		_, err = db.Exec(createUsersTable)
		if err != nil {
			log.Fatal("Failed to create users table:", err)
		}
		log.Println("Created users table")
	} else if err != nil {
		log.Fatal("Error checking users table:", err)
	} else {
		// Table exists, check and add missing columns
		checkAndAddColumn("users", "email", "VARCHAR(255)")
		checkAndAddColumn("users", "password_hash", "TEXT DEFAULT ''")
		checkAndAddColumn("users", "is_guest", "BOOLEAN DEFAULT FALSE")
		checkAndAddColumn("users", "created_at", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP")
		checkAndAddColumn("users", "rule", "INTEGER DEFAULT 1")

		// Set default for existing rows with NULL password_hash
		_, err = db.Exec("UPDATE users SET password_hash = '' WHERE password_hash IS NULL")
		if err != nil {
			log.Printf("Warning: Could not update NULL password_hash: %v", err)
		}

		// Try to set default constraint if not already set
		_, err = db.Exec("ALTER TABLE users ALTER COLUMN password_hash SET DEFAULT ''")
		if err != nil {
			log.Printf("Warning: Could not set default for password_hash: %v", err)
		}
	}

	// Insert default guest user (with empty password_hash)
	insertGuest := `
	INSERT INTO users (username, email, password_hash, is_guest, rule)
	SELECT 'guest', 'guest@tytyber.ru', '', TRUE, 1
	WHERE NOT EXISTS (
		SELECT 1 FROM users WHERE username = 'guest'
	);`

	_, err = db.Exec(insertGuest)
	if err != nil {
		log.Fatal("Failed to insert guest user:", err)
	}

	// Insert default admin user (for admin panel)
	insertAdmin := `
	INSERT INTO users (username, email, password_hash, is_guest, rule)
	SELECT 'admin', 'admin@tytyber.ru', '', FALSE, 3
	WHERE NOT EXISTS (
		SELECT 1 FROM users WHERE username = 'admin'
	);`

	_, err = db.Exec(insertAdmin)
	if err != nil {
		log.Fatal("Failed to insert admin user:", err)
	}

	// Insert default moderator user
	insertModerator := `
	INSERT INTO users (username, email, password_hash, is_guest, rule)
	SELECT 'moderator', 'moderator@tytyber.ru', '', FALSE, 2
	WHERE NOT EXISTS (
		SELECT 1 FROM users WHERE username = 'moderator'
	);`

	_, err = db.Exec(insertModerator)
	if err != nil {
		log.Fatal("Failed to insert moderator user:", err)
	}

	log.Println("Database tables initialized successfully")
}

func checkAndAddColumn(table, column, definition string) {
	var colName string
	err := db.QueryRow("SELECT column_name FROM information_schema.columns WHERE table_name = $1 AND column_name = $2", table, column).Scan(&colName)

	if err == sql.ErrNoRows {
		_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s", table, column, definition))
		if err != nil {
			log.Fatalf("Failed to add %s column: %v", column, err)
		}
		log.Printf("Added %s column to %s table", column, table)
	}
}

func GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := db.QueryRow("SELECT id, username, email, password_hash, is_guest, rule, created_at FROM users WHERE username = $1", username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.IsGuest,
		&user.Rule,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func CreateUser(username, password string) error {
	_, err := db.Exec("INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)", username, "", password)
	return err
}

func CheckUserCredentials(username, password string) (bool, error) {
	var storedPassword string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&storedPassword)
	if err != nil {
		return false, err
	}
	return storedPassword == password, nil
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}

// GetSessionUserFromRequest returns the current session user from the request
// It reads from current_user cookie for persistence across requests
func GetSessionUserFromRequest(r *http.Request) (*models.User, error) {
	state := GetSessionState()

	// First check if user is available in session state
	if state.User != nil {
		return state.User, nil
	}

	// Try to get user from cookie directly
	if r != nil {
		if cookie, err := r.Cookie("current_user"); err == nil {
			user, err := GetUserByUsername(cookie.Value)
			if err == nil {
				return user, nil
			}
		}
	}

	// Return guest if no user found
	return GetUserByUsername("guest")
}

// GetUserByUsernameByID gets user by ID
func GetUserByUsernameByID(userID int) (*models.User, error) {
	user := &models.User{}
	err := db.QueryRow("SELECT id, username, email, password_hash, is_guest, rule, created_at FROM users WHERE id = $1", userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.IsGuest,
		&user.Rule,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// SetSessionUserFromRequest sets the current user in the session for the request
func SetSessionUserFromRequest(w http.ResponseWriter, r *http.Request, user *models.User) error {
	// Set cookie directly for immediate availability
	http.SetCookie(w, &http.Cookie{
		Name:     "current_user",
		Value:    user.Username,
		Path:     "/",
		MaxAge:   3600 * 24 * 7, // 7 days
		HttpOnly: true,
		Secure:   false,
	})
	return nil
}

// ClearSessionUserFromRequest clears the current session for the request
func ClearSessionUserFromRequest(w http.ResponseWriter, r *http.Request) error {
	http.SetCookie(w, &http.Cookie{
		Name:     "current_user",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Expire immediately
		HttpOnly: true,
		Secure:   false,
	})
	return nil
}

// StartSessionCleanup starts a goroutine that clears session every hour
func StartSessionCleanup() {
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			sessionState = nil
		}
	}()
}

func GetSessionState() *models.SessionState {
	if sessionState == nil {
		sessionState = &models.SessionState{}
	}
	return sessionState
}

func SetSessionState(state *models.SessionState) {
	sessionState = state
}

// GetTotalUsers returns total number of users
func GetTotalUsers() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// GetNewUsersToday returns number of users registered today
func GetNewUsersToday() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= date_trunc('day', now())").Scan(&count)
	return count, err
}

// GetDailyVisits returns daily visits count
func GetDailyVisits() (int, error) {
	var count int
	// Placeholder - would need visits tracking table
	count = 150
	return count, nil
}

// GetMonthlyVisits returns monthly visits count
func GetMonthlyVisits() (int, error) {
	var count int
	// Placeholder - would need visits tracking table
	count = 4500
	return count, nil
}

// GetActiveUsers returns active users count
func GetActiveUsers() (int, error) {
	var count int
	// Placeholder - would need session tracking
	count = 25
	return count, nil
}
