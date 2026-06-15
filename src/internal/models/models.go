package models

import "time"

type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"-" db:"password_hash"`
	Email     string    `json:"email" db:"email"`
	IsGuest   bool      `json:"is_guest" db:"is_guest"`
	Rule      int       `json:"rule" db:"rule"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Message struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Rule      int       `json:"rule" db:"rule"`
	Message   string    `json:"message" db:"message"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type TerminalRequest struct {
	Command   string `json:"command"`
	Timestamp string `json:"timestamp"`
}

type TerminalResponse struct {
	Output             string   `json:"output"`
	Type               string   `json:"type"`
	Additional         []string `json:"additional,omitempty"`
	CurrentSessionUser string   `json:"current_session_user,omitempty"`
	UpdateCookie       string   `json:"update_cookie,omitempty"`
	WaitForInput       bool     `json:"wait_for_input,omitempty"`
	Redirect           string   `json:"redirect,omitempty"`
}

type UserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	User    *User  `json:"user,omitempty"`
}

type DarkModeResponse struct {
	Enabled bool `json:"enabled"`
}

type SessionState struct {
	Active     bool
	StateType  string // "register_username", "register_password", "register_repeat", "login_username", "login_password"
	Username   string
	Password   string
	User       *User // stores full user object with rule info
}

type BlogPost struct {
	ID        int       `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	ImageURL  string    `json:"image_url" db:"image_url"`
	ImageFile string    `json:"image_file" db:"image_file"`
	Author    string    `json:"author" db:"author"`
	Views     int       `json:"views" db:"views"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
