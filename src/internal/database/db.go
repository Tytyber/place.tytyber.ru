package database

import (
	"database/sql"
	"fmt"
	"log"
	"place-tytyber/internal/models"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB(host string, port int, user string, password string, dbname string) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Successfully connected to database")

	// Create tables and add missing columns
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
	INSERT INTO users (username, email, password_hash, is_guest)
	SELECT 'guest', 'guest@tytyber.ru', '', TRUE
	WHERE NOT EXISTS (
		SELECT 1 FROM users WHERE username = 'guest'
	);`

	_, err = db.Exec(insertGuest)
	if err != nil {
		log.Fatal("Failed to insert guest user:", err)
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
	err := db.QueryRow("SELECT id, username, email, password_hash, is_guest, created_at FROM users WHERE username = $1", username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.IsGuest,
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

// GetSessionUser returns the current session user (simulated)
var currentSessionUser *models.User

// Session state for interactive commands
var sessionState *models.SessionState

func SetSessionUser(user *models.User) {
	currentSessionUser = user
}

func GetSessionUser() *models.User {
	// If no user in session, return guest
	if currentSessionUser == nil {
		if guest, err := GetUserByUsername("guest"); err == nil {
			return guest
		}
	}
	return currentSessionUser
}

func ClearSessionUser() {
	currentSessionUser = nil
	sessionState = nil
}

func UpdateSessionUser(username string) error {
	user, err := GetUserByUsername(username)
	if err != nil {
		return err
	}
	currentSessionUser = user
	return nil
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

// StartSessionCleanup starts a goroutine that clears session every hour
func StartSessionCleanup() {
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			ClearSessionUser()
		}
	}()
}
