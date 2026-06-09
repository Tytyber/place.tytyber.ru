package main

import (
	"fmt"
	"log"
	"net/http"
	"place-tytyber/internal/database"
	"place-tytyber/internal/handlers"
)

func main() {
	// Initialize database
	// You can change these values to connect to your PostgreSQL
	database.InitDB("192.168.1.3", 5432, "postgres", "", "tytyber_club")

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

	// Serve static files from templates directory
	http.HandleFunc("/assets/", handlers.HandleStaticFiles)

	fmt.Println("Server starting on :8080")
	fmt.Println("Database initialized successfully")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
