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

func HandleAdminPanel(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	user, err := database.GetSessionUserFromRequest(r)
	if err != nil {
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	// Check if user has admin rights (rule >= 2)
	if user.Rule < 2 {
		http.Error(w, "Access denied. Admin panel is only available for moderators and administrators.", http.StatusForbidden)
		return
	}

	// Determine rule level name
	ruleName := "Модератор"
	if user.Rule == 3 {
		ruleName = "Администратор"
	}

	// Get statistics from database
	totalUsers, err := database.GetTotalUsers()
	if err != nil {
		totalUsers = 0
	}

	newUsersToday, err := database.GetNewUsersToday()
	if err != nil {
		newUsersToday = 0
	}

	dailyVisits, err := database.GetDailyVisits()
	if err != nil {
		dailyVisits = 0
	}

	monthlyVisits, err := database.GetMonthlyVisits()
	if err != nil {
		monthlyVisits = 0
	}

	activeUsers, err := database.GetActiveUsers()
	if err != nil {
		activeUsers = 0
	}

	// Prepare template data
	data := map[string]interface{}{
		"TotalUsers":        totalUsers,
		"NewUsersToday":     newUsersToday,
		"DailyVisits":       dailyVisits,
		"MonthlyVisits":     monthlyVisits,
		"ActiveUsers":       activeUsers,
		"CurrentUsername":   user.Username,
		"CurrentRule":       user.Rule,
		"CurrentRuleName":   ruleName,
		"OnlinePercentage":  0,
		"NewUsers":          0,
		"ReturningUsers":    0,
		"InactiveUsers":     0,
		"PeakVisits":        0,
		"AvgVisits":         0,
		"CurrentUsers":      0,
		"TargetUsers":       100,
		"GrowthPercent":     0,
		"NewSignups":        0,
		"ReturningCount":    0,
		"MobileUsers":       0,
		"Countries":         0,
		"AiInsight":         "Аналитика генерируется...",
		"Uptime":            99,
		"UptimeChange":      1,
		"AvgDailyVisits":    0,
		"PreviousAvg":       0,
		"VisitsChange":      0,
		"MonthlyGrowth":     0,
		"SignupsCount":      0,
		"ReturningSessions": 0,
		"MobilePercent":     0,
		"CountriesGrowth":   0,
	}

	// Serve adminPanel.html template
	err = templates.ExecuteTemplate(w, "adminPanel.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func HandleCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Return current session user
	user, _ := database.GetSessionUserFromRequest(r)
	
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
