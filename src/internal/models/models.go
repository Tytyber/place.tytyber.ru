package models

import "time"

type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"-" db:"password_hash"`
	Email     string    `json:"email" db:"email"`
	IsGuest   bool      `json:"is_guest" db:"is_guest"`
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
}
