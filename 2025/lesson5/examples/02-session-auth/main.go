package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// User - структура пользователя
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"` // Хеш пароля, не отправляется клиенту
	Name     string `json:"name"`
}

// Хранилище пользователей (в реальности - БД)
var users = map[string]*User{
	"user@example.com": {
		ID:       1,
		Email:    "user@example.com",
		Password: hashPassword("password123"),
		Name:     "Иван Иванов",
	},
}

// Хранилище сессий (в продакшене - Redis)
var sessions = make(map[string]int) // session_id -> user_id
var sessionsMu sync.RWMutex

// Хеширование пароля
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// Генерация случайного session ID
func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Handler регистрации
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Проверка, что пользователь не существует
	if _, exists := users[req.Email]; exists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Создание пользователя
	user := &User{
		ID:       len(users) + 1,
		Email:    req.Email,
		Password: hashPassword(req.Password),
		Name:     req.Name,
	}
	users[req.Email] = user

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Handler логина
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Проверка пользователя
	user, exists := users[req.Email]
	if !exists || user.Password != hashPassword(req.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Создание сессии
	sessionID := generateSessionID()
	sessionsMu.Lock()
	sessions[sessionID] = user.ID
	sessionsMu.Unlock()

	// Установка cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600 * 24, // 24 часа
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Logged in successfully",
		"user":    user,
	})
}

// Handler выхода
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionsMu.Lock()
		delete(sessions, cookie.Value)
		sessionsMu.Unlock()
	}

	// Удаление cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// Middleware проверки аутентификации
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		sessionsMu.RLock()
		userID, exists := sessions[cookie.Value]
		sessionsMu.RUnlock()

		if !exists {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		// Добавляем userID в контекст
		ctx := context.WithValue(r.Context(), "userID", userID)
		next(w, r.WithContext(ctx))
	}
}

// Получение пользователя по ID
func getUserByID(id int) *User {
	for _, user := range users {
		if user.ID == id {
			return user
		}
	}
	return nil
}

// Защищенный handler - профиль пользователя
func profileHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	user := getUserByID(userID)

	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func main() {
	// Публичные endpoints
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)

	// Защищенные endpoints
	http.HandleFunc("/profile", authMiddleware(profileHandler))

	fmt.Println("Server started on :8080")
	fmt.Println("\nTry:")
	fmt.Println("  # Регистрация")
	fmt.Println(`  curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"test123","name":"Test User"}'`)
	fmt.Println("\n  # Логин (сохраните cookie в файл)")
	fmt.Println(`  curl -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"email":"user@example.com","password":"password123"}' -c cookies.txt`)
	fmt.Println("\n  # Доступ к профилю с cookie")
	fmt.Println("  curl http://localhost:8080/profile -b cookies.txt")
	fmt.Println("\n  # Выход")
	fmt.Println("  curl -X POST http://localhost:8080/logout -b cookies.txt")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
