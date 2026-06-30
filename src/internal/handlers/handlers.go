package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"place-tytyber/internal/database"
	"place-tytyber/internal/models"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
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
	"seq": func(n int) []int {
		result := make([]int, n)
		for i := range result {
			result[i] = i + 1
		}
		return result
	},
	"substr": func(s string, start, length int) string {
		if start >= len(s) {
			return ""
		}
		end := start + length
		if end > len(s) {
			end = len(s)
		}
		return s[start:end]
	},
	"isImage": func(url string) bool {
		if url == "" {
			return false
		}
		lower := strings.ToLower(url)
		return strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".gif") || strings.HasSuffix(lower, ".webp")
	},
	"markdown": func(text string) template.HTML {
		p := parser.New()
		h := html.NewRenderer(html.RendererOptions{})
		return template.HTML(markdown.ToHTML([]byte(text), p, h))
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

	// Get forum statistics
	forumStats, err := database.GetForumStats()
	if err != nil {
		forumStats = &models.ForumStats{}
	}

	dailyForumTopics, err := database.GetDailyForumTopics()
	if err != nil {
		dailyForumTopics = 0
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
		// Forum statistics
		"TotalForumTopics":  forumStats.TotalTopics,
		"TotalForumReplies": forumStats.TotalReplies,
		"TotalForumViews":   forumStats.TotalViews,
		"DailyForumTopics":  dailyForumTopics,
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
		"Users":       users,
		"TotalCount":  totalCount,
		"TotalPages":  totalPages,
		"CurrentPage": page,
		"SearchTerm":  searchTerm,
		"CurrentRule": currentUser.Rule,
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
		"Messages":    messages,
		"CurrentUser": user,
		"CurrentRule": user.Rule,
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

// Blog handlers

func HandleBlogList(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/blog" {
		http.NotFound(w, r)
		return
	}

	// Get page number
	page := 1
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	// Get blog posts
	posts, err := database.GetBlogPosts(page, 10)
	if err != nil {
		http.Error(w, "Error getting blog posts", http.StatusInternalServerError)
		return
	}

	// Get total count and calculate pages
	totalCount, err := database.GetBlogPostCount()
	if err != nil {
		totalCount = 0
	}
	totalPages := (totalCount + 9) / 10

	// Prepare template data
	data := map[string]interface{}{
		"Posts":       posts,
		"TotalPages":  totalPages,
		"CurrentPage": page,
	}

	// Serve blogList.html template
	err = templates.ExecuteTemplate(w, "blogList.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func HandleBlogPost(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path
	path := r.URL.Path
	if len(path) <= 10 || path[:10] != "/blog/post" {
		http.NotFound(w, r)
		return
	}

	// Get ID after /blog/post/
	idStr := path[11:]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Increment view count
	database.IncrementBlogPostViews(id)

	// Get blog post
	post, err := database.GetBlogPostByID(id)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"Post": post,
	}

	// Serve blogPost.html template
	err = templates.ExecuteTemplate(w, "blogPost.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func HandleBlogAdmin(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	user, err := database.GetSessionUserFromRequest(r)
	if err != nil {
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	// Check if user has admin rights (rule == 3)
	if user.Rule != 3 {
		http.Error(w, "Access denied. Blog admin is only available for administrators.", http.StatusForbidden)
		return
	}

	// Get page number
	page := 1
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	// Get blog posts
	posts, err := database.GetBlogPosts(page, 10)
	if err != nil {
		http.Error(w, "Error getting blog posts", http.StatusInternalServerError)
		return
	}

	// Get total count and calculate pages
	totalCount, err := database.GetBlogPostCount()
	if err != nil {
		totalCount = 0
	}
	totalPages := (totalCount + 9) / 10

	// Prepare template data
	data := map[string]interface{}{
		"Posts":       posts,
		"TotalPages":  totalPages,
		"CurrentPage": page,
		"CurrentUser": user,
	}

	// Serve blogAdmin.html template
	err = templates.ExecuteTemplate(w, "blogAdmin.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func HandleCreateBlogPost(w http.ResponseWriter, r *http.Request) {
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

	// Check if user has admin rights (rule == 3)
	if user.Rule != 3 {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse multipart form data (for file upload)
	err = r.ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	imageUrl := r.FormValue("image_url")
	imageFile := ""

	// Handle file upload
	if file, header, err := r.FormFile("image_file"); err == nil {
		defer file.Close()

		// Validate file type
		if !isImageFile(header.Filename) {
			http.Error(w, "Invalid file type. Only images are allowed", http.StatusBadRequest)
			return
		}

		// Save file to uploads directory
		filename, err := saveUploadedFile(file, header)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		imageFile = filename
	}

	if title == "" || content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	// Create blog post
	err = database.CreateBlogPost(title, content, user.Username, imageUrl, imageFile)
	if err != nil {
		http.Error(w, "Error creating blog post", http.StatusInternalServerError)
		return
	}

	// Redirect back to blog admin
	http.Redirect(w, r, "/admin/blog", http.StatusSeeOther)
}

func HandleEditBlogPost(w http.ResponseWriter, r *http.Request) {
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

	// Check if user has admin rights (rule == 3)
	if user.Rule != 3 {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse multipart form data (for file upload)
	err = r.ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	imageUrl := r.FormValue("image_url")
	imageFile := r.FormValue("image_file")

	// Handle new file upload
	if file, header, err := r.FormFile("new_image_file"); err == nil {
		defer file.Close()

		// Validate file type
		if !isImageFile(header.Filename) {
			http.Error(w, "Invalid file type. Only images are allowed", http.StatusBadRequest)
			return
		}

		// Save file to uploads directory
		filename, err := saveUploadedFile(file, header)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		imageFile = filename
	}

	if title == "" || content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	// Update blog post
	err = database.UpdateBlogPost(postID, title, content, imageUrl, imageFile)
	if err != nil {
		http.Error(w, "Error updating blog post", http.StatusInternalServerError)
		return
	}

	// Redirect back to blog admin
	http.Redirect(w, r, "/admin/blog", http.StatusSeeOther)
}

func HandleCreateBlogPostForm(w http.ResponseWriter, r *http.Request) {
	// Get current user from session
	user, err := database.GetSessionUserFromRequest(r)
	if err != nil {
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	// Check if user has admin rights (rule == 3)
	if user.Rule != 3 {
		http.Error(w, "Access denied. Blog admin is only available for administrators.", http.StatusForbidden)
		return
	}

	// Serve blogCreate.html template
	err = templates.ExecuteTemplate(w, "blogCreate.html", nil)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func HandleDeleteBlogPost(w http.ResponseWriter, r *http.Request) {
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

	// Check if user has admin rights (rule == 3)
	if user.Rule != 3 {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Delete blog post
	err = database.DeleteBlogPost(postID)
	if err != nil {
		http.Error(w, "Error deleting blog post", http.StatusInternalServerError)
		return
	}

	// Redirect back to blog admin
	http.Redirect(w, r, "/admin/blog", http.StatusSeeOther)
}

// isImageFile checks if the file is an image
func isImageFile(filename string) bool {
	lower := strings.ToLower(filename)
	return strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".gif") || strings.HasSuffix(lower, ".webp")
}

// saveUploadedFile saves the uploaded file to the uploads directory
func saveUploadedFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Create uploads directory if it doesn't exist
	uploadsDir := "./internal/templates/uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return "", err
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s", timestamp, filepath.Base(header.Filename))

	// Create destination file
	destPath := filepath.Join(uploadsDir, filename)
	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return filename, nil
}

func HandleUploadsFiles(w http.ResponseWriter, r *http.Request) {
	// Build file path for uploads directory
	if r.URL.Path == "/uploads/" || r.URL.Path == "/uploads" {
		http.NotFound(w, r)
		return
	}

	var filePath string
	if len(r.URL.Path) > 9 && r.URL.Path[:9] == "/uploads/" {
		// Serve from uploads directory inside templates
		filePath = filepath.Join("./internal/templates", "uploads", r.URL.Path[9:])
	} else {
		http.NotFound(w, r)
		return
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

func HandleProfilePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/profile" {
		http.NotFound(w, r)
		return
	}

	// Get current user from session
	user, err := database.GetSessionUserFromRequest(r)
	if err != nil {
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	// Get user statistics
	messagesCount, err := database.GetMessagesCount()
	if err != nil {
		messagesCount = 0
	}

	postsCount, err := database.GetBlogPostCount()
	if err != nil {
		postsCount = 0
	}

	totalViews, err := database.GetTotalViews()
	if err != nil {
		totalViews = 0
	}

	// Determine rule name
	ruleName := "Обычный пользователь"
	if user.Rule == 2 {
		ruleName = "Модератор"
	} else if user.Rule == 3 {
		ruleName = "Администратор"
	}

	// Prepare template data
	data := map[string]interface{}{
		"Username":      user.Username,
		"ID":            user.ID,
		"Email":         user.Email,
		"Rule":          user.Rule,
		"RuleName":      ruleName,
		"CreatedAt":     user.CreatedAt,
		"Avatar":        user.Avatar,
		"MessagesCount": messagesCount,
		"PostsCount":    postsCount,
		"TotalViews":    totalViews,
	}

	// Set no-cache headers to prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Serve profile.html template
	err = templates.ExecuteTemplate(w, "profile.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Clear session
	database.ClearSessionUserFromRequest(w, r)

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get current user from session
	currentUser, err := database.GetSessionUserFromRequest(r)
	if err != nil {
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	// Parse multipart form data (for file upload) or regular form data
	err = r.ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		// If multipart fails, try regular form parsing
		if err = r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}
	}

	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid user ID: %s", r.FormValue("user_id")), http.StatusBadRequest)
		return
	}

	// Check if user is authorized to update this profile
	if currentUser.ID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")

	// Handle avatar upload
	avatarFile := ""
	if file, header, err := r.FormFile("avatar"); err == nil {
		defer file.Close()

		// Validate file type
		if !isImageFile(header.Filename) {
			http.Error(w, "Invalid file type. Only images are allowed", http.StatusBadRequest)
			return
		}

		// Save file to uploads directory
		filename, err := saveUploadedFile(file, header)
		if err != nil {
			http.Error(w, "Error saving avatar", http.StatusInternalServerError)
			return
		}
		avatarFile = filename
	}

	// Update user profile
	err = database.UpdateUserProfile(userID, username, email, avatarFile)
	if err != nil {
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	// Update session user if username changed
	if username != currentUser.Username {
		if user, err := database.GetUserByUsername(username); err == nil {
			database.SetSessionUserFromRequest(w, r, user)
		}
	}

	// Redirect back to profile
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// Forum handlers

func HandleForum(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/forum" && !strings.HasPrefix(r.URL.Path, "/forum/") {
		http.NotFound(w, r)
		return
	}

	// Get current user from session
	currentUser, err := database.GetSessionUserFromRequest(r)
	if err != nil {
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	// Get folder ID from URL
	folderID := 0
	if len(r.URL.Path) > 7 && strings.HasPrefix(r.URL.Path, "/forum/f/") {
		folderStr := strings.TrimPrefix(r.URL.Path, "/forum/f/")
		if f, err := strconv.Atoi(folderStr); err == nil {
			folderID = f
		}
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

	// Get topics
	topics, totalCount, err := database.GetForumTopics(page, 10, folderID, searchTerm)
	if err != nil {
		http.Error(w, "Error getting topics", http.StatusInternalServerError)
		return
	}

	// Calculate total pages
	totalPages := (totalCount + 9) / 10

	// Get all folders
	folders, _ := database.GetAllFolders()

	// Get current folder name
	currentFolderName := "Все темы"
	if folderID > 0 {
		if folder, err := database.GetFolderByID(folderID); err == nil {
			currentFolderName = folder.Name
		}
	}

	// Prepare template data
	data := map[string]interface{}{
		"Topics":            topics,
		"Folders":           folders,
		"CurrentFolder":     folderID,
		"CurrentFolderName": currentFolderName,
		"TotalPages":        totalPages,
		"CurrentPage":       page,
		"SearchTerm":        searchTerm,
		"CurrentUser":       currentUser,
		"CurrentRule":       currentUser.Rule,
	}

	// Serve forum.html template
	err = templates.ExecuteTemplate(w, "forum.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func HandleForumTopic(w http.ResponseWriter, r *http.Request) {
	// Extract topic ID from URL path (StripPrefix removed the /forum/topic/ prefix)
	idStr := r.URL.Path
	if idStr == "" {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Increment view count
	database.IncrementTopicViews(id)

	// Get forum topic
	topic, err := database.GetForumTopicByID(id)
	if err != nil {
		http.Error(w, "Topic not found", http.StatusNotFound)
		return
	}

	// Get replies
	page := 1
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	replies, replyCount, err := database.GetTopicReplies(id, page, 20)
	if err != nil {
		http.Error(w, "Error getting replies", http.StatusInternalServerError)
		return
	}

	totalPages := (replyCount + 19) / 20

	// Get current user
	currentUser, err := database.GetSessionUserFromRequest(r)
	if err != nil {
		http.Error(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	// Check if user is author
	isAuthor := false
	if currentUser.ID != 0 {
		isAuthor, _ = database.IsTopicAuthor(id, currentUser.ID)
	}

	// Prepare template data
	data := map[string]interface{}{
		"Topic":       topic,
		"Replies":     replies,
		"TotalPages":  totalPages,
		"CurrentPage": page,
		"CurrentUser": currentUser,
		"CurrentRule": currentUser.Rule,
		"IsAuthor":    isAuthor,
	}

	// Serve forumTopic.html template
	err = templates.ExecuteTemplate(w, "forumTopic.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func HandleCreateTopic(w http.ResponseWriter, r *http.Request) {
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

	// Check if user is registered (rule >= 1 and not guest)
	if user.IsGuest {
		http.Error(w, "Only registered users can create topics", http.StatusForbidden)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	folderID := 0
	if folderParam := r.FormValue("folder_id"); folderParam != "" {
		if f, err := strconv.Atoi(folderParam); err == nil {
			folderID = f
		}
	}

	if title == "" || content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	// Create topic
	err = database.CreateForumTopic(title, content, folderID, user.Username, user.ID)
	if err != nil {
		http.Error(w, "Error creating topic", http.StatusInternalServerError)
		return
	}

	// Redirect to forum
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

func HandleCreateReply(w http.ResponseWriter, r *http.Request) {
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

	// Check if user is registered (rule >= 1 and not guest)
	if user.IsGuest {
		http.Error(w, "Only registered users can reply", http.StatusForbidden)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	topicID, err := strconv.Atoi(r.FormValue("topic_id"))
	if err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Create reply
	err = database.CreateForumReply(topicID, content, user.Username, user.ID)
	if err != nil {
		http.Error(w, "Error creating reply", http.StatusInternalServerError)
		return
	}

	// Update topic replies count and last active
	database.UpdateTopicReplies(topicID)
	database.UpdateTopicLastActive(topicID)

	// Redirect back to topic
	http.Redirect(w, r, "/forum/topic/"+strconv.Itoa(topicID), http.StatusSeeOther)
}

func HandleDeleteTopic(w http.ResponseWriter, r *http.Request) {
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

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	topicID, err := strconv.Atoi(r.FormValue("topic_id"))
	if err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	// Check permissions: author or rule >= 2
	isAuthor, err := database.IsTopicAuthor(topicID, currentUser.ID)
	if err != nil {
		http.Error(w, "Error checking permissions", http.StatusInternalServerError)
		return
	}

	if !isAuthor && currentUser.Rule < 2 {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Delete topic
	err = database.DeleteForumTopic(topicID)
	if err != nil {
		http.Error(w, "Error deleting topic", http.StatusInternalServerError)
		return
	}

	// Redirect back to forum
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

func HandleDeleteReply(w http.ResponseWriter, r *http.Request) {
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

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	replyID, err := strconv.Atoi(r.FormValue("reply_id"))
	if err != nil {
		http.Error(w, "Invalid reply ID", http.StatusBadRequest)
		return
	}

	// Check permissions: author or rule >= 2
	isAuthor, err := database.IsReplyAuthor(replyID, currentUser.ID)
	if err != nil {
		http.Error(w, "Error checking permissions", http.StatusInternalServerError)
		return
	}

	if !isAuthor && currentUser.Rule < 2 {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get topic ID before deleting
	topicID, err := database.GetTopicIDByReplyID(replyID)
	if err != nil {
		http.Error(w, "Error getting topic ID", http.StatusInternalServerError)
		return
	}

	// Delete reply
	err = database.DeleteForumReply(replyID)
	if err != nil {
		http.Error(w, "Error deleting reply", http.StatusInternalServerError)
		return
	}

	// Redirect back to topic
	http.Redirect(w, r, "/forum/topic/"+strconv.Itoa(topicID), http.StatusSeeOther)
}

func HandleCreateFolder(w http.ResponseWriter, r *http.Request) {
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

	// Check if user has admin rights (rule >= 2)
	if currentUser.Rule < 2 {
		http.Error(w, "Access denied. Only moderators and administrators can create folders.", http.StatusForbidden)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	slug := r.FormValue("slug")
	description := r.FormValue("description")
	icon := r.FormValue("icon")
	order, err := strconv.Atoi(r.FormValue("order"))
	if err != nil {
		order = 0
	}

	if name == "" || slug == "" {
		http.Error(w, "Name and slug are required", http.StatusBadRequest)
		return
	}

	// Create folder
	err = database.CreateForumFolder(name, slug, description, icon, order)
	if err != nil {
		http.Error(w, "Error creating folder", http.StatusInternalServerError)
		return
	}

	// Redirect to forum
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

func HandleDeleteFolder(w http.ResponseWriter, r *http.Request) {
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

	// Check if user has admin rights (rule >= 2)
	if currentUser.Rule < 2 {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	folderID, err := strconv.Atoi(r.FormValue("folder_id"))
	if err != nil {
		http.Error(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	// Delete folder
	err = database.DeleteForumFolder(folderID)
	if err != nil {
		http.Error(w, "Error deleting folder", http.StatusInternalServerError)
		return
	}

	// Redirect back to forum
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}
