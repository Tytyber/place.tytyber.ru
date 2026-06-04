# Tytyber Place - Cyberpunk Terminal Web Application

![Tytyber](https://img.shields.io/badge/Tytyber-Place-ccff00?style=for-the-badge&logo=terminal)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=for-the-badge&logo=postgresql)

> The future isn't minimal. It's systematic.

## 🚀 Overview

Tytyber Place is a cyberpunk-themed web application with an interactive terminal interface. Built with Go and PostgreSQL, it features a dark mode experience where users can register, login, and interact with a system terminal directly from their browser.

## ✨ Features

- **Interactive Terminal** - Full-featured command-line interface in the browser
- **User Authentication** - Register new users and login securely
- **Dark Mode** - Cyberpunk-themed dark interface for enhanced experience
- **Session Management** - Persistent user sessions across page reloads
- **PostgreSQL Backend** - Reliable database for user storage
- **Real-time Updates** - Instant feedback for all terminal commands

## 🛠️ Technology Stack

- **Backend**: Go 1.21+
- **Database**: PostgreSQL 15+
- **Frontend**: HTML5, CSS3, JavaScript (ES6+)
- **API**: RESTful JSON API

## 📋 Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Git

## 🚀 Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/place.tytyber.ru.git
cd place.tytyber.ru
```

### 2. Setup Database

1. Create a new PostgreSQL database:

```sql
CREATE DATABASE tytyber_club;
```

2. The application will automatically create the `users` table on first run

### 3. Configure Database Connection

Edit the database connection settings in `src/main.go`:

```go
database.InitDB("192.168.1.12", 5432, "postgres", "", "tytyber_club")
```

Change the following values:
- Host: Your PostgreSQL host (default: `192.168.1.12`)
- Port: PostgreSQL port (default: `5432`)
- User: Your PostgreSQL username (default: `postgres`)
- Password: Your PostgreSQL password (default: empty)
- Database: Your database name (default: `tytyber_club`)

### 4. Build and Run

```bash
# Initialize Go modules
go mod tidy

# Build the application
go build -o bin/server.exe

# Run the server
./bin/server.exe
```

The server will start on `http://localhost:8080`

## 📖 Usage

### Main Commands

Once the server is running, open your browser and navigate to `http://localhost:8080`

#### Terminal Commands

| Command | Description |
|---------|-------------|
| `register` | Register a new user (interactive) |
| `login` | Login with existing credentials (interactive) |
| `logout` | Logout and return to guest mode |
| `dark-mode` | Enable dark mode |
| `light-mode` | Switch to light mode |
| `dark` | Show dark mode information |
| `help` | Show available commands |
| `clear` | Clear terminal output |
| `test` | Test server connection |

#### Interactive Commands

**Register a new user:**
```
register
# Follow prompts to enter username, password, and confirm password
```

**Login:**
```
login
# Follow prompts to enter username and password
```

### Dark Mode

To access dark mode:
1. Open the terminal in the main page
2. Type `dark-mode` and press Enter
3. The app will redirect you to `/dark-mode` with full dark theme

## 📁 Project Structure

```
place.tytyber.ru/
├── src/
│   ├── internal/
│   │   ├── database/       # Database operations
│   │   │   └── db.go
│   │   ├── handlers/       # HTTP handlers
│   │   │   ├── handlers.go
│   │   │   └── terminal.go
│   │   ├── models/         # Data models
│   │   │   └── models.go
│   │   └── templates/      # HTML templates
│   │       ├── index.html
│   │       ├── darkMode.html
│   │       └── assets/
│   │           ├── js/
│   │           │   └── index.js
│   │           └── styles/
│   │               ├── index.css
│   │               └── dark.css
│   └── main.go
├── go.mod
├── go.sum
└── README.md
```

## 🔐 Security

- Passwords are stored in plain text in the database (for simplicity in this demo)
- Session management uses cookies
- Dark mode is protected and requires terminal access

## 🐛 Known Issues

- Password validation is case-sensitive
- Session cleanup runs every hour

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License.

## 🙏 Acknowledgments

- Tytyber Community
- Cyberpunk theme inspiration
- All contributors and users

## 📞 Contact

- Discord: [Join Tytyber Club](https://discord.gg/yjB6pMrVYx)
- Telegram: [@tytyber_club](https://t.me/tytyber_club)
- GitHub: [@Tytyber](https://github.com/Tytyber)

---

> **Access Granted** _© 2026 Tytyber Club Team. ALL RIGHTS RESERVED._
