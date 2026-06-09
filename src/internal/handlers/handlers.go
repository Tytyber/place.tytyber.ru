package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"place-tytyber/internal/database"
	"place-tytyber/internal/models"
)

// Global template cache
var templates *template.Template

// Template functions
var templateFuncs = template.FuncMap{
	"sub": func(a, b int) int {
		return a - b
	},
	"add": func(a, b int) int {
		return a + b
	},
}

func init() {
	var err error
	templates, err = template.New("").Funcs(templateFuncs).ParseGlob("./internal/templates/*.html")
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

	monthlyNewUsers, err := database.GetMonthlyNewUsers()
	if err != nil {
		monthlyNewUsers = 0
	}

	totalVisits, err := database.GetTotalVisits()
	if err != nil {
		totalVisits = 0
	}

	monthlyVisits, err := database.GetMonthlyVisits()
	if err != nil {
		monthlyVisits = 0
	}

	activeUsers, err := database.GetActiveUsers()
	if err != nil {
		activeUsers = 0
	}

	dailyVisits, err := database.GetDailyVisits()
	if err != nil {
		dailyVisits = 0
	}

	// Calculate online percentage
	var onlinePercentage float64
	if totalUsers > 0 {
		onlinePercentage = float64(activeUsers) / float64(totalUsers) * 100
	}

	// Get uptime data
	uptime, err := database.GetUptime()
	if err != nil {
		uptime = 99
	}

	uptimeChange, err := database.GetUptimeChange()
	if err != nil {
		uptimeChange = 1
	}

	// Prepare template data
	data := map[string]interface{}{
		"TotalUsers":        totalUsers,
		"MonthlyNewUsers":   monthlyNewUsers,
		"TotalVisits":       totalVisits,
		"MonthlyVisits":     monthlyVisits,
		"DailyVisits":       dailyVisits,
		"ActiveUsers":       activeUsers,
		"CurrentUsername":   user.Username,
		"CurrentRule":       user.Rule,
		"CurrentRuleName":   ruleName,
		"OnlinePercentage":  int(onlinePercentage),
		"Uptime":            uptime,
		"UptimeChange":      uptimeChange,
		"NewUsers":          monthlyNewUsers,
		"ReturningUsers":    totalUsers - monthlyNewUsers,
		"InactiveUsers":     0,
		"CurrentUsers":      totalUsers,
		"TargetUsers":       totalUsers + 50,
		"GrowthPercent":     0,
		"NewSignups":        monthlyNewUsers,
		"ReturningCount":    totalUsers,
		"MobileUsers":       totalUsers / 2,
		"Countries":         5,
		"AiInsight":         "Система показывает стабильный рост пользовательской базы. Рекомендуется уделить внимание вовлечению новых пользователей.",
		"SignupsCount":      monthlyNewUsers,
		"ReturningSessions": totalUsers,
		"MobilePercent":     50,
		"CountriesGrowth":   1,
	}

	// Serve adminPanel.html template
	err = templates.ExecuteTemplate(w, "adminPanel.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func HandleUsersPage(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	currentUser, err := database.GetSessionUserFromRequest(r)
	if err != nil {
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	// Check if user has admin rights (rule == 3)
	if currentUser.Rule != 3 {
		http.Error(w, "Access denied. Users page is only available for administrators.", http.StatusForbidden)
		return
	}

	// Get search term
	searchTerm := r.URL.Query().Get("search")
	
	// Get page number
	page := 1
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	// Get users
	var users []*models.User
	var totalCount int
	if searchTerm != "" {
		users, err = database.SearchUsers(searchTerm)
		totalCount = len(users)
	} else {
		users, err = database.GetAllUsers(page, 20)
		totalCount, err = database.GetUserCount()
	}

	if err != nil {
		http.Error(w, "Error getting users", http.StatusInternalServerError)
		return
	}

	// Calculate total pages
	totalPages := (totalCount + 19) / 20

	// Prepare template data
	data := map[string]interface{}{
		"Users":        users,
		"TotalCount":   totalCount,
		"TotalPages":   totalPages,
		"CurrentPage":  page,
		"SearchTerm":   searchTerm,
		"CurrentRule":  currentUser.Rule,
	}

	// Serve users.html template
	err = templates.ExecuteTemplate(w, "users.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	// Get current user from session
	currentUser, err := database.GetSessionUserFromRequest(r)
	if err != nil {
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	// Check if user has admin rights (rule == 3)
	if currentUser.Rule != 3 {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	rule, err := strconv.Atoi(r.FormValue("rule"))
	if err != nil {
		rule = 1
	}

	// Update user
	err = database.UpdateUser(userID, email, rule)
	if err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	// Redirect back to users page
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	// Get current user from session
	currentUser, err := database.GetSessionUserFromRequest(r)
	if err != nil {
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	// Check if user has admin rights (rule == 3)
	if currentUser.Rule != 3 {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Don't allow deleting yourself
	if currentUser.ID == userID {
		http.Error(w, "Cannot delete yourself", http.StatusBadRequest)
		return
	}

	// Delete user
	err = database.DeleteUser(userID)
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	// Redirect back to users page
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func HandleAdminChat(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	user, err := database.GetSessionUserFromRequest(r)
	if err != nil {
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	// Check if user has admin rights (rule >= 2)
	if user.Rule < 2 {
		http.Error(w, "Access denied. Admin chat is only available for moderators and administrators.", http.StatusForbidden)
		return
	}

	// Get messages
	messages, err := database.GetMessages(50)
	if err != nil {
		http.Error(w, "Error getting messages", http.StatusInternalServerError)
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"Messages":     messages,
		"CurrentUser":  user,
		"CurrentRule":  user.Rule,
	}

	// Serve adminChat.html template
	err = templates.ExecuteTemplate(w, "adminChat.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func HandleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	// Get current user from session
	user, err := database.GetSessionUserFromRequest(r)
	if err != nil {
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	// Check if user has admin rights (rule >= 2)
	if user.Rule < 2 {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	message := r.FormValue("message")
	if message == "" {
		http.Redirect(w, r, "/admin/chat", http.StatusSeeOther)
		return
	}

	// Add message
	err = database.AddMessage(user.Username, user.Rule, message)
	if err != nil {
		http.Error(w, "Error adding message", http.StatusInternalServerError)
		return
	}

	// Redirect back to chat
	http.Redirect(w, r, "/admin/chat", http.StatusSeeOther)
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
