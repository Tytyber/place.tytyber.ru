package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"place-tytyber/internal/database"
	"place-tytyber/internal/models"
	"strings"
)

func HandleTerminal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.TerminalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.TerminalResponse{
			Output: "Все работает",
			Type:   "success",
		})
		return
	}

	// Check if we're in dark mode from cookie
	darkModeEnabled := false
	if cookie, err := r.Cookie("dark_mode"); err == nil && cookie.Value == "true" {
		darkModeEnabled = true
	}
	
	// Check if we have an active session state (interactive command)
	state := database.GetSessionState()
	if state.Active {
		response := processInteractiveCommand(req.Command, darkModeEnabled)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Process command
	response := processCommand(req.Command, darkModeEnabled)

	// Update cookie if dark mode state changed
	if response.UpdateCookie != "" {
		http.SetCookie(w, &http.Cookie{
			Name:  "dark_mode",
			Value: response.UpdateCookie,
			Path:  "/",
		})
	}
	
	// Update session user using gorilla sessions
	if response.CurrentSessionUser != "" {
		if user, err := database.GetUserByUsername(response.CurrentSessionUser); err == nil {
			database.SetSessionUserFromRequest(w, r, user)
		}
	}

	// Clear session if logout
	if response.Type == "success" && response.CurrentSessionUser == "guest" {
		database.ClearSessionUserFromRequest(w, r)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func processCommand(cmd string, darkModeEnabled bool) models.TerminalResponse {
	cmd = strings.ToLower(strings.TrimSpace(cmd))

	switch cmd {
	case "ping":
		return models.TerminalResponse{
			Output: "Все работает",
			Type:   "success",
		}

	case "light-mode":
		if darkModeEnabled {
			return models.TerminalResponse{
				Output:       "Возвращение к обычному режиму...",
				Type:         "success",
				UpdateCookie: "false",
				Additional: []string{
					"Переход на главную страницу...",
				},
			}
		} else {
			return models.TerminalResponse{
				Output: "Уже в светлом режиме",
				Type:   "info",
			}
		}

	case "dark-mode":
		state := database.GetSessionState()
		return models.TerminalResponse{
			Output:       "Переход к dark mode...",
			Type:         "success",
			UpdateCookie: "true",
			CurrentSessionUser: state.Username,
			Additional: []string{
				"Следующая команда отправит вас на /dark-mode",
				"Ожидание подтверждения...",
			},
		}

	case "dark":
		status := "Неактивен"
		if darkModeEnabled {
			status = "Активен"
		}
		return models.TerminalResponse{
			Output: fmt.Sprintf("Dark mode доступен по адресу: /dark-mode\nТекущее состояние: %s", status),
			Type:   "info",
			Additional: []string{
				"Доступные команды:",
				"  dark-mode     - показать информацию о dark mode",
				"  light-mode   - переключить на светлый режим",
				"  light        - переключить на светлый режим",
				"  dark         - информация о dark mode",
				"  test         - проверка соединения",
				"  register     - регистрация нового пользователя",
				"  login        - авторизация пользователя",
				"  logout       - выход из системы",
				"  clear        - очистить терминал",
				"  help         - список команд",
			},
		}

	case "help":
		return models.TerminalResponse{
			Output: "Доступные команды:",
			Type:   "info",
			Additional: []string{
				"  ping         - проверка соединения",
				"  register     - регистрация нового пользователя",
				"  login        - авторизация пользователя",
				"  logout       - выход из системы",
				"  clear        - очистить терминал",
				"  help         - показать это сообщение",
			},
		}

	case "clear":
		return models.TerminalResponse{
			Output:       "clear",
			Type:         "clear",
			UpdateCookie: "",
		}

	case "register":
		return startRegister()

	case "login":
		return startLogin()

	case "logout":
		return processLogout()

	default:
		return processCommandDefault(cmd)
	}
}

func processCommandDefault(cmd string) models.TerminalResponse {
	return models.TerminalResponse{
		Output: fmt.Sprintf("Unknown command: %s", cmd),
		Type:   "error",
	}
}

// processInteractiveCommand handles commands during an interactive session
func processInteractiveCommand(input string, darkModeEnabled bool) models.TerminalResponse {
	state := database.GetSessionState()
	input = strings.TrimSpace(input)

	switch state.StateType {
	case "register_username":
		return handleRegisterUsername(input)
	case "register_password":
		state.Password = input
		state.StateType = "register_repeat"
		return models.TerminalResponse{
			Output:       "Повторите пароль:",
			Type:         "info",
			WaitForInput: true,
		}
	case "register_repeat":
		return handleRegisterRepeat(input, state.Username, state.Password)
	case "login_username":
		state.Username = input
		state.StateType = "login_password"
		return models.TerminalResponse{
			Output:       "Введите пароль:",
			Type:         "info",
			WaitForInput: true,
		}
	case "login_password":
		return handleLoginPassword(input, state.Username)
	default:
		// Reset state if unknown
		state.Active = false
		state.StateType = ""
		return processCommandDefault(input)
	}
}

func startRegister() models.TerminalResponse {
	state := database.GetSessionState()
	state.Active = true
	state.StateType = "register_username"
	state.Username = ""
	state.Password = ""

	return models.TerminalResponse{
		Output:       "Введите имя пользователя:",
		Type:         "info",
		WaitForInput: true,
	}
}

func handleRegisterUsername(username string) models.TerminalResponse {
	state := database.GetSessionState()
	state.Username = username
	// Password will be stored in next step
	state.StateType = "register_password"

	return models.TerminalResponse{
		Output:       "Введите пароль:",
		Type:         "info",
		WaitForInput: true,
	}
}

func handleRegisterRepeat(repeatPassword, username, password string) models.TerminalResponse {
	state := database.GetSessionState()

	if password != repeatPassword {
		state.Active = false
		state.StateType = ""
		return models.TerminalResponse{
			Output: "Ошибка: пароли не совпадают. Регистрация отменена.",
			Type:   "error",
			Additional: []string{
				"Введите 'register', чтобы начать заново",
			},
		}
	}

	// Check if user already exists
	user, err := database.GetUserByUsername(username)
	if err == nil && user != nil {
		state.Active = false
		state.StateType = ""
		return models.TerminalResponse{
			Output: fmt.Sprintf("Ошибка: пользователь '%s' уже существует", username),
			Type:   "error",
			Additional: []string{
				"Введите 'register', чтобы начать заново",
			},
		}
	}

	// Create user
	err = database.CreateUser(username, password)
	if err != nil {
		state.Active = false
		state.StateType = ""
		return models.TerminalResponse{
			Output: fmt.Sprintf("Ошибка при создании пользователя: %v", err),
			Type:   "error",
			Additional: []string{
				"Введите 'register', чтобы начать заново",
			},
		}
	}

	state.Active = false
	state.StateType = ""
	return models.TerminalResponse{
		Output:             fmt.Sprintf("Пользователь '%s' успешно зарегистрирован!", username),
		Type:               "success",
		CurrentSessionUser: username,
		Additional:         []string{
			"Теперь вы можете авторизоваться с помощью команды:",
			"  login",
		},
	}
}

func startLogin() models.TerminalResponse {
	state := database.GetSessionState()
	state.Active = true
	state.StateType = "login_username"
	state.Username = ""
	state.Password = ""

	return models.TerminalResponse{
		Output:       "Введите имя пользователя:",
		Type:         "info",
		WaitForInput: true,
	}
}

func handleLoginPassword(password, username string) models.TerminalResponse {
	state := database.GetSessionState()

	// Check credentials
	if username == "guest" {
		state.Active = false
		state.StateType = ""
		
		return models.TerminalResponse{
			Output:             "Вы вошли как гость",
			Type:               "success",
			CurrentSessionUser: "guest",
			Additional: []string{
				"Смена приглашения терминала...",
				"Следующая команда обновит видимость",
			},
		}
	}

	// Check if user exists and credentials are correct
	isValid, err := database.CheckUserCredentials(username, password)
	if err != nil {
		state.Active = false
		state.StateType = ""
		return models.TerminalResponse{
			Output: "Ошибка при проверке учетных данных",
			Type:   "error",
			Additional: []string{
				fmt.Sprintf("Детали: %v", err),
			},
		}
	}

	if !isValid {
		state.Active = false
		state.StateType = ""
		return models.TerminalResponse{
			Output: "Ошибка: неверное имя пользователя или пароль",
			Type:   "error",
			Additional: []string{
				"Введите 'login', чтобы попробовать снова",
			},
		}
	}

	state.Active = false
	state.StateType = ""
	return models.TerminalResponse{
		Output:             fmt.Sprintf("Добро пожаловать, %s!", username),
		Type:               "success",
		CurrentSessionUser: username,
		Additional: []string{
			"Смена приглашения терминала...",
			"Следующая команда обновит видимость",
		},
	}
}

func processLogout() models.TerminalResponse {
	state := database.GetSessionState()
	state.Active = false
	state.StateType = ""
	state.Username = ""
	state.Password = ""
	
	return models.TerminalResponse{
		Output:             "Вы вышли из системы",
		Type:               "success",
		CurrentSessionUser: "guest",
		Additional: []string{
			"Смена приглашения терминала...",
			"Теперь вы вошли как гость",
		},
	}
}
