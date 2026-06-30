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
			Output:             "Переход к dark mode...",
			Type:               "success",
			UpdateCookie:       "true",
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
				"  profile      - перейти на страницу профиля",
				"  admin-panel  - перейти в админ-панель",
				"  forum        - перейти на форум",
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
				"  profile      - перейти на страницу профиля",
				"  admin-panel  - перейти в админ-панель",
				"  forum        - перейти на форум",
				"  clear        - очистить терминал",
				"  help         - показать это сообщение",
			},
		}

	case "clear":
		// Clear should also reset to guest if not logged in
		state := database.GetSessionState()
		return models.TerminalResponse{
			Output:             "clear",
			Type:               "clear",
			UpdateCookie:       "",
			CurrentSessionUser: state.Username,
		}

	case "register":
		return startRegister()

	case "login":
		return startLogin()

	case "logout":
		return processLogout()

	case "admin-panel":
		return processAdminPanel()

	case "profile":
		return processProfile()

	case "forum":
		return processForum()

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
	state.User = nil

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
		Additional: []string{
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
	state.User = nil

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

		// Get guest user object with rule info
		guestUser, err := database.GetUserByUsername("guest")
		if err == nil {
			state.User = guestUser
		}

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

	// Get full user object with rule info
	user, err := database.GetUserByUsername(username)
	if err != nil {
		state.Active = false
		state.StateType = ""
		return models.TerminalResponse{
			Output: "Ошибка при получении информации о пользователе",
			Type:   "error",
			Additional: []string{
				fmt.Sprintf("Детали: %v", err),
			},
		}
	}

	// Save user object to session state for later use
	state.User = user

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
	state.User = nil

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

func processAdminPanel() models.TerminalResponse {
	state := database.GetSessionState()

	// Use user from session state if available
	var user *models.User
	if state.User != nil {
		user = state.User
	} else {
		return models.TerminalResponse{
			Output: "Ошибка: информация о пользователе недоступна. Пожалуйста, войдите в систему.",
			Type:   "error",
		}
	}

	// Проверяем права доступа (только модераторы и администраторы)
	if user.Rule < 2 {
		ruleName := "Обычный пользователь"
		if user.Rule == 1 {
			ruleName = "Обычный пользователь"
		}
		return models.TerminalResponse{
			Output: "Доступ запрещен. Для доступа к админ-панели необходимы права модератора или администратора.",
			Type:   "error",
			Additional: []string{
				"Войдите как пользователь с правами rule >= 2",
				"Текущий пользователь:",
				fmt.Sprintf("  Имя: %s", user.Username),
				fmt.Sprintf("  Права: %d (%s)", user.Rule, ruleName),
			},
		}
	}

	// Определяем уровень прав
	ruleName := "Модератор"
	if user.Rule == 3 {
		ruleName = "Администратор"
	}

	return models.TerminalResponse{
		Output:             "Переход в админ-панель...",
		Type:               "success",
		CurrentSessionUser: user.Username,
		Redirect:           "/admin",
		Additional: []string{
			"Проверка прав доступа...",
			fmt.Sprintf("Уровень прав: %d - %s", user.Rule, ruleName),
			"Админ-панель доступна по адресу: /admin",
			"Ожидание перенаправления...",
		},
	}
}

func processProfile() models.TerminalResponse {
	state := database.GetSessionState()

	// Use user from session state if available
	var user *models.User
	if state.User != nil {
		user = state.User
	} else {
		return models.TerminalResponse{
			Output: "Ошибка: информация о пользователе недоступна. Пожалуйста, войдите в систему.",
			Type:   "error",
		}
	}

	// Определяем уровень прав
	ruleName := "Обычный пользователь"
	if user.Rule == 2 {
		ruleName = "Модератор"
	} else if user.Rule == 3 {
		ruleName = "Администратор"
	}

	return models.TerminalResponse{
		Output:             "Переход на страницу профиля...",
		Type:               "success",
		CurrentSessionUser: user.Username,
		Redirect:           "/profile",
		Additional: []string{
			"Получение информации о пользователе...",
			fmt.Sprintf("Имя: %s", user.Username),
			fmt.Sprintf("Права: %d - %s", user.Rule, ruleName),
			"Страница профиля доступна по адресу: /profile",
			"Ожидание перенаправления...",
		},
	}
}

func processForum() models.TerminalResponse {
	state := database.GetSessionState()

	// Use user from session state if available
	var user *models.User
	if state.User != nil {
		user = state.User
	} else {
		// Get user by username if available
		if state.Username != "" {
			var err error
			user, err = database.GetUserByUsername(state.Username)
			if err != nil {
				// Fallback to guest if user not found
				return models.TerminalResponse{
					Output:             "Переход на форум...",
					Type:               "success",
					CurrentSessionUser: "guest",
					Redirect:           "/forum",
					Additional: []string{
						"Форум доступен по адресу: /forum",
						"Возможности:",
						"  Просмотр тем в папках",
						"  Поиск по заголовкам и пользователям",
						"  Создание тем (только для зарегистрированных)",
						"  Ответы на темы",
						"  Управление папками (для прав >= 2)",
						"Ожидание перенаправления...",
					},
				}
			}
		}
	}

	// If we have a valid user, use their information
	if user != nil {
		return models.TerminalResponse{
			Output:             "Переход на форум...",
			Type:               "success",
			CurrentSessionUser: user.Username,
			Redirect:           "/forum",
			Additional: []string{
				"Форум доступен по адресу: /forum",
				"Возможности:",
				"  Просмотр тем в папках",
				"  Поиск по заголовкам и пользователям",
				"  Создание тем (только для зарегистрированных)",
				"  Ответы на темы",
				"  Управление папками (для прав >= 2)",
				"Ожидание перенаправления...",
			},
		}
	}

	// Fallback to guest if no user found
	return models.TerminalResponse{
		Output:             "Переход на форум...",
		Type:               "success",
		CurrentSessionUser: "guest",
		Redirect:           "/forum",
		Additional: []string{
			"Форум доступен по адресу: /forum",
			"Возможности:",
			"  Просмотр тем в папках",
			"  Поиск по заголовкам и пользователям",
			"  Создание тем (только для зарегистрированных)",
			"  Ответы на темы",
			"  Управление папками (для прав >= 2)",
			"Ожидание перенаправления...",
		},
	}
}
