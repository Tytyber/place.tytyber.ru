# Использование Gorilla Sessions для управления сессиями

## Изменения в проекте

### 1. Добавление зависимости
В `go.mod` добавлена зависимость `github.com/gorilla/sessions v1.4.0`

### 2. Изменения в `internal/database/db.go`

#### Инициализация store:
```go
store = sessions.NewCookieStore([]byte("place-tytyber-secret-key-change-in-production"))
```

#### Новые функции:
- `GetSessionUserFromRequest(r *http.Request)` - получает текущего пользователя из сессии
- `SetSessionUserFromRequest(w http.ResponseWriter, r *http.Request, user *models.User)` - устанавливает пользователя в сессию
- `ClearSessionUserFromRequest(w http.ResponseWriter, r *http.Request)` - очищает сессию
- `GetUserByUsernameByID(userID int)` - получает пользователя по ID

#### Удалены:
- `RestoreSessionFromCookie()` - больше не нужна, так как sessions управляются автоматически
- `GetSessionUser()` - заменена на `GetSessionUserFromRequest()`

### 3. Изменения в `internal/handlers/handlers.go`

#### `HandleCurrentUser`:
Теперь использует `database.GetSessionUserFromRequest(r)` вместо ручного чтения cookie.

### 4. Изменения в `internal/handlers/terminal.go`

#### `HandleTerminal`:
- Удалено чтение cookie `current_user`
- Установка сессии теперь происходит через `database.SetSessionUserFromRequest()`
- Очистка сессии при логауте через `database.ClearSessionUserFromRequest()`

## Преимущества Gorilla Sessions

1. **Автоматическое управление**: Сессии автоматически сохраняются и восстанавливаются
2. **Безопасность**: HttpOnly cookie предотвращает доступ к сессии из JavaScript
3. **Глобальные сессии**: Сессия доступна на всех страницах через `store.Get(r, "session-name")`
4. **Надежность**: Не зависит от ручного управления cookie
5. **Гибкость**: Легко изменить время жизни сессии, путь и другие параметры

## Как это работает

1. При входе пользователя через `/api/terminal` вызывается `SetSessionUserFromRequest()`
2. Gorilla sessions создает cookie с ID сессии
3. При любом запросе к серверу вызывается `store.Get(r, "place-tytyber-session")`
4. Sessions автоматически загружаются из cookie
5. Данные из сессии (user_id, username) доступны через `session.Values`
6. При логауте вызывается `ClearSessionUserFromRequest()` для очистки сессии

## Миграция с ручного управления cookie

**Было:**
```go
// Установка cookie
http.SetCookie(w, &http.Cookie{
    Name:  "current_user",
    Value: username,
    Path:  "/",
})

// Получение cookie
if cookie, err := r.Cookie("current_user"); err == nil {
    currentUser = cookie.Value
}
```

**Стало:**
```go
// Установка сессии
database.SetSessionUserFromRequest(w, r, user)

// Получение из сессии
user, err := database.GetSessionUserFromRequest(r)
```

## Дополнительные настройки

В `db.go` можно изменить параметры сессии:
```go
store = sessions.NewCookieStore([]byte("secret-key"))
store.Options = &sessions.Options{
    Path:     "/",
    MaxAge:   3600 * 24 * 7, // 7 дней
    HttpOnly: true,
    Secure:   false, // true для HTTPS
    SameSite: http.SameSiteLaxMode,
}
```
