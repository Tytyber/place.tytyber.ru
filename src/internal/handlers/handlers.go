package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"place-tytyber/internal/database"
)

// Global template cache
var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseGlob("./internal/templates/*.html")
	if err != nil {
		panic("Failed to parse templates: " + err.Error())
	}
}

func HandleMainPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Serve index.html template
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func HandleDarkModePage(w http.ResponseWriter, r *http.Request) {
	// Check if user came from terminal with dark_mode cookie set
	cookie, err := r.Cookie("dark_mode")
	if err != nil || cookie.Value != "true" {
		http.Error(w, "Access denied. Use terminal to enable dark mode.", http.StatusForbidden)
		return
	}

	// Serve darkMode.html template
	err = templates.ExecuteTemplate(w, "darkMode.html", nil)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func HandleCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Return current session user
	user := database.GetSessionUser()
	
	response := map[string]string{
		"username": "guest",
	}
	
	if user != nil {
		response["username"] = user.Username
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleStaticFiles(w http.ResponseWriter, r *http.Request) {
	// Build file path - handle assets directory separately
	var filePath string
	if r.URL.Path == "/assets/" || r.URL.Path == "/assets" {
		http.NotFound(w, r)
		return
	}

	if len(r.URL.Path) > 8 && r.URL.Path[:8] == "/assets/" {
		// Serve from assets directory inside templates
		filePath = filepath.Join("./internal/templates", "assets", r.URL.Path[8:])
	} else {
		// Serve from templates directory
		filePath = filepath.Join("./internal/templates", r.URL.Path)
	}

	// Check if file exists
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Serve the file
	http.ServeFile(w, r, filePath)
}
