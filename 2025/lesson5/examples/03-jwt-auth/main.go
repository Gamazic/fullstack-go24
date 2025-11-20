package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("super-secret-key-change-in-production")

// User - структура пользователя
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Name     string `json:"name"`
}

// Хранилище пользователей
var users = map[string]*User{
	"user@example.com": {
		ID:       1,
		Email:    "user@example.com",
		Password: hashPassword("password123"),
		Name:     "Иван Иванов",
	},
}

// Хеширование пароля
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// Создание JWT токена
func createToken(userID int, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
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

	// Проверка существования пользователя
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

	// Создание токена
	token, err := createToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"token":   token,
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

	// Создание JWT токена
	token, err := createToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Logged in successfully",
		"token":   token,
		"user":    user,
	})
}

// Middleware проверки JWT
func jwtAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Формат: "Bearer TOKEN"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Парсинг и проверка токена
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Проверка метода подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Извлечение claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Добавляем данные пользователя в контекст
		ctx := context.WithValue(r.Context(), "userID", int(claims["user_id"].(float64)))
		ctx = context.WithValue(ctx, "email", claims["email"].(string))
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
	email := r.Context().Value("email").(string)

	user := getUserByID(userID)
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":  user,
		"email": email,
	})
}

// Handler для проверки токена
func verifyHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	email := r.Context().Value("email").(string)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":   true,
		"user_id": userID,
		"email":   email,
	})
}

func main() {
	// Публичные endpoints
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)

	// Защищенные endpoints
	http.HandleFunc("/profile", jwtAuthMiddleware(profileHandler))
	http.HandleFunc("/verify", jwtAuthMiddleware(verifyHandler))

	fmt.Println("Server started on :8080")
	fmt.Println("\nTry:")
	fmt.Println("  # Логин (получить токен)")
	fmt.Println(`  TOKEN=$(curl -s -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"email":"user@example.com","password":"password123"}' | jq -r '.token')`)
	fmt.Println("\n  # Доступ к профилю с токеном")
	fmt.Println(`  curl http://localhost:8080/profile -H "Authorization: Bearer $TOKEN"`)
	fmt.Println("\n  # Проверка токена")
	fmt.Println(`  curl http://localhost:8080/verify -H "Authorization: Bearer $TOKEN"`)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
