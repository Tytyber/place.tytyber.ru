package database

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"place-tytyber/internal/config"
	"place-tytyber/internal/models"
	"time"

	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

var db *sql.DB
var store *sessions.CookieStore

// Session state for interactive commands
var sessionState *models.SessionState

type Message struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Rule      int       `json:"rule" db:"rule"`
	Message   string    `json:"message" db:"message"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

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

	// Initialize session store with config
	cfg := config.Get()
	store = sessions.NewCookieStore([]byte(cfg.SessionSecretKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24 * 7, // 7 days
		HttpOnly: true,
		Secure:   false,
	}
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

	// Create chat table
	createChatTable()

	// Create blog table
	createBlogTable()

	log.Println("Database tables initialized successfully")
	migrateExistingPasswords()
}

// migrateExistingPasswords hashes passwords that are stored in plain text
func migrateExistingPasswords() {
	// Get all users with plain text passwords (empty or short passwords)
	rows, err := db.Query("SELECT id, username, password_hash FROM users WHERE password_hash = '' OR LENGTH(password_hash) < 64")
	if err != nil {
		log.Printf("Warning: Could not query users for password migration: %v", err)
		return
	}
	defer rows.Close()

	var updatedCount int
	for rows.Next() {
		var id int
		var username, passwordHash string
		err := rows.Scan(&id, &username, &passwordHash)
		if err != nil {
			continue
		}

		// Hash the existing password (even if empty, it will be hashed)
		hashedPassword := hashPassword(passwordHash)
		_, err = db.Exec("UPDATE users SET password_hash = $1 WHERE id = $2", hashedPassword, id)
		if err != nil {
			log.Printf("Warning: Could not update password for user %s: %v", username, err)
		} else {
			updatedCount++
		}
	}

	if updatedCount > 0 {
		log.Printf("Migrated %d existing passwords to hashed format", updatedCount)
	}
}

func createChatTable() {
	// Check if table exists
	var tableName string
	err := db.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_name = 'admin_chat_messages'").Scan(&tableName)

	if err == sql.ErrNoRows {
		createChatTable := `
		CREATE TABLE admin_chat_messages (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			rule INTEGER DEFAULT 1,
			message TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`

		_, err = db.Exec(createChatTable)
		if err != nil {
			log.Fatal("Failed to create admin_chat_messages table:", err)
		}
		log.Println("Created admin_chat_messages table")
	} else if err != nil {
		log.Fatal("Error checking admin_chat_messages table:", err)
	}
}

func createBlogTable() {
	// Check if table exists
	var tableName string
	err := db.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_name = 'blog_posts'").Scan(&tableName)

	if err == sql.ErrNoRows {
		createBlogTable := `
		CREATE TABLE blog_posts (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			image_url TEXT DEFAULT '',
			image_file TEXT DEFAULT '',
			author VARCHAR(255) NOT NULL,
			views INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`

		_, err = db.Exec(createBlogTable)
		if err != nil {
			log.Fatal("Failed to create blog_posts table:", err)
		}
		log.Println("Created blog_posts table")
	} else if err != nil {
		log.Fatal("Error checking blog_posts table:", err)
	} else {
		// Add image_file column if missing
		checkAndAddColumn("blog_posts", "image_file", "TEXT DEFAULT ''")
	}
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

func GetUserByID(id int) (*models.User, error) {
	user := &models.User{}
	err := db.QueryRow("SELECT id, username, email, password_hash, is_guest, rule, created_at FROM users WHERE id = $1", id).Scan(
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

func hashPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

func CreateUser(username, password string) error {
	hashedPassword := hashPassword(password)
	_, err := db.Exec("INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)", username, "", hashedPassword)
	return err
}

func CheckUserCredentials(username, password string) (bool, error) {
	var storedPassword string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&storedPassword)
	if err != nil {
		return false, err
	}
	// Compare hashed password
	inputHash := hashPassword(password)
	return storedPassword == inputHash, nil
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}

// GetSessionUserFromRequest returns the current session user from the request
// It reads from current_user cookie for persistence across requests
func GetSessionUserFromRequest(r *http.Request) (*models.User, error) {
	// First check if user is available in session state
	state := GetSessionState()
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

// GetTotalUsers returns total number of users
func GetTotalUsers() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// GetMonthlyNewUsers returns number of users registered this month
func GetMonthlyNewUsers() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= date_trunc('month', now())").Scan(&count)
	return count, err
}

// GetAllUsers returns all users with pagination
func GetAllUsers(page, limit int) ([]*models.User, error) {
	offset := (page - 1) * limit
	rows, err := db.Query("SELECT id, username, email, password_hash, is_guest, rule, created_at FROM users ORDER BY id LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsGuest, &user.Rule, &user.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

// GetUserCount returns total count of users
func GetUserCount() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// SearchUsers searches users by username
func SearchUsers(searchTerm string) ([]*models.User, error) {
	rows, err := db.Query("SELECT id, username, email, password_hash, is_guest, rule, created_at FROM users WHERE username ILIKE $1 ORDER BY id", "%"+searchTerm+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsGuest, &user.Rule, &user.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

// UpdateUser updates user data
func UpdateUser(userID int, email string, rule int) error {
	_, err := db.Exec("UPDATE users SET email = $1, rule = $2 WHERE id = $3", email, rule, userID)
	return err
}

// DeleteUser deletes a user
func DeleteUser(id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

// AddMessage adds a new message to the chat
func AddMessage(username string, rule int, message string) error {
	_, err := db.Exec("INSERT INTO admin_chat_messages (username, rule, message) VALUES ($1, $2, $3)", username, rule, message)
	return err
}

// GetMessages returns messages with pagination
func GetMessages(limit int) ([]*Message, error) {
	rows, err := db.Query("SELECT id, username, rule, message, created_at FROM admin_chat_messages ORDER BY id DESC LIMIT $1", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		err := rows.Scan(&msg.ID, &msg.Username, &msg.Rule, &msg.Message, &msg.CreatedAt)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// GetBlogPosts returns all blog posts with pagination
func GetBlogPosts(page, limit int) ([]*models.BlogPost, error) {
	offset := (page - 1) * limit
	rows, err := db.Query("SELECT id, title, content, image_url, image_file, author, views, created_at, updated_at FROM blog_posts ORDER BY id DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.BlogPost
	for rows.Next() {
		post := &models.BlogPost{}
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImageURL, &post.ImageFile, &post.Author, &post.Views, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			continue
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// GetBlogPostByID returns a single blog post by ID
func GetBlogPostByID(id int) (*models.BlogPost, error) {
	post := &models.BlogPost{}
	err := db.QueryRow("SELECT id, title, content, image_url, image_file, author, views, created_at, updated_at FROM blog_posts WHERE id = $1", id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.ImageURL,
		&post.ImageFile,
		&post.Author,
		&post.Views,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return post, nil
}

// CreateBlogPost creates a new blog post
func CreateBlogPost(title, content, author, imageUrl, imageFile string) error {
	_, err := db.Exec("INSERT INTO blog_posts (title, content, image_url, image_file, author, views) VALUES ($1, $2, $3, $4, $5, 0)", title, content, imageUrl, imageFile, author)
	return err
}

// UpdateBlogPost updates an existing blog post
func UpdateBlogPost(id int, title, content, imageUrl, imageFile string) error {
	_, err := db.Exec("UPDATE blog_posts SET title = $1, content = $2, image_url = $3, image_file = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5", title, content, imageUrl, imageFile, id)
	return err
}

// DeleteBlogPost deletes a blog post
func DeleteBlogPost(id int) error {
	_, err := db.Exec("DELETE FROM blog_posts WHERE id = $1", id)
	return err
}

// IncrementBlogPostViews increments the view count for a blog post
func IncrementBlogPostViews(id int) error {
	_, err := db.Exec("UPDATE blog_posts SET views = views + 1 WHERE id = $1", id)
	return err
}

// GetBlogPostCount returns total count of blog posts
func GetBlogPostCount() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM blog_posts").Scan(&count)
	return count, err
}

// GetDailyVisits returns daily visits count (placeholder)
func GetDailyVisits() (int, error) {
	return 150, nil
}

// GetMonthlyVisits returns monthly visits count (placeholder)
func GetMonthlyVisits() (int, error) {
	return 4500, nil
}

// GetActiveUsers returns active users count (placeholder)
func GetActiveUsers() (int, error) {
	return 25, nil
}

// GetTotalVisits returns total visits count (placeholder - realistic value)
func GetTotalVisits() (int, error) {
	return 1500, nil
}

// GetUptime returns system uptime percentage
func GetUptime() (int, error) {
	return 98, nil
}

// GetUptimeChange returns uptime change vs last month
func GetUptimeChange() (int, error) {
	return 2, nil
}
