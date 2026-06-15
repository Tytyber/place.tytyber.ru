package main

import (
	"fmt"
	"log"
	"net/http"
	"place-tytyber/internal/config"
	"place-tytyber/internal/database"
	"place-tytyber/internal/handlers"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	database.InitDB(
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseName,
	)

	// Start session cleanup
	database.StartSessionCleanup()

	// Terminal API endpoint
	http.HandleFunc("/api/terminal", handlers.HandleTerminal)

	// Current user API endpoint
	http.HandleFunc("/api/current-user", handlers.HandleCurrentUser)

	// Dark mode page handler (only accessible via /dark-mode)
	http.HandleFunc("/dark-mode", handlers.HandleDarkModePage)

	// Admin panel page handler (only accessible for users with rule >= 2)
	http.HandleFunc("/admin", handlers.HandleAdminPanel)
	http.HandleFunc("/admin/", handlers.HandleAdminPanel)

	// Users page handler (only accessible for users with rule == 3)
	http.HandleFunc("/admin/users", handlers.HandleUsersPage)

	// Update user handler
	http.HandleFunc("/admin/users/update", handlers.HandleUpdateUser)

	// Delete user handler
	http.HandleFunc("/admin/users/delete", handlers.HandleDeleteUser)

	// Admin chat page handler (only accessible for users with rule >= 2)
	http.HandleFunc("/admin/chat", handlers.HandleAdminChat)
	http.HandleFunc("/admin/chat/", handlers.HandleAdminChat)

	// Send message handler
	http.HandleFunc("/admin/chat/send", handlers.HandleSendMessage)

	// Main page handler
	http.HandleFunc("/", handlers.HandleMainPage)

	// Blog handlers
	http.HandleFunc("/blog", handlers.HandleBlogList)
	http.HandleFunc("/blog/post/", handlers.HandleBlogPost)

	// Admin blog handlers
	http.HandleFunc("/admin/blog", handlers.HandleBlogAdmin)
	http.HandleFunc("/admin/blog/create", handlers.HandleCreateBlogPost)
	http.HandleFunc("/admin/blog/create-form", handlers.HandleCreateBlogPostForm)
	http.HandleFunc("/admin/blog/edit", handlers.HandleEditBlogPost)
	http.HandleFunc("/admin/blog/delete", handlers.HandleDeleteBlogPost)

	// Serve static files from templates directory
	http.HandleFunc("/assets/", handlers.HandleStaticFiles)
	http.HandleFunc("/uploads/", handlers.HandleUploadsFiles)

	fmt.Printf("Server starting on %s:%d\n", cfg.ServerHost, cfg.ServerPort)
	fmt.Println("Database initialized successfully")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.ServerPort), nil))
}
