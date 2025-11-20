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

var jwtSecret = []byte("super-secret-key")

// Role - тип для ролей
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
	RoleModerator Role = "moderator"
)

// User - структура пользователя
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Name     string `json:"name"`
	Role     Role   `json:"role"`
}

// Хранилище пользователей
var users = map[string]*User{
	"admin@example.com": {
		ID:       1,
		Email:    "admin@example.com",
		Password: hashPassword("admin123"),
		Name:     "Администратор",
		Role:     RoleAdmin,
	},
	"moderator@example.com": {
		ID:       2,
		Email:    "moderator@example.com",
		Password: hashPassword("mod123"),
		Name:     "Модератор",
		Role:     RoleModerator,
	},
	"user@example.com": {
		ID:       3,
		Email:    "user@example.com",
		Password: hashPassword("user123"),
		Name:     "Обычный пользователь",
		Role:     RoleUser,
	},
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// Создание JWT токена с ролью
func createToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    string(user.Role),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
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

	user, exists := users[req.Email]
	if !exists || user.Password != hashPassword(req.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := createToken(user)
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
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		ctx := context.WithValue(r.Context(), "userID", int(claims["user_id"].(float64)))
		ctx = context.WithValue(ctx, "email", claims["email"].(string))
		ctx = context.WithValue(ctx, "role", Role(claims["role"].(string)))

		next(w, r.WithContext(ctx))
	}
}

// Middleware проверки конкретной роли
func requireRole(role Role, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value("role").(Role)

		if userRole != role {
			http.Error(w, fmt.Sprintf("Forbidden: requires %s role", role), http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// Middleware проверки одной из ролей
func requireAnyRole(roles []Role, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value("role").(Role)

		for _, role := range roles {
			if userRole == role {
				next(w, r)
				return
			}
		}

		http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
	}
}

// Публичный endpoint
func publicHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"message": "This is a public endpoint",
	})
}

// Endpoint для любого аутентифицированного пользователя
func profileHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	email := r.Context().Value("email").(string)
	role := r.Context().Value("role").(Role)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"email":   email,
		"role":    role,
	})
}

// Endpoint только для администраторов
func adminHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Welcome to admin panel",
		"users":   getAllUsers(),
	})
}

// Endpoint для администраторов и модераторов
func moderationHandler(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value("role").(Role)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Welcome to moderation panel",
		"role":    role,
	})
}

// Endpoint для удаления пользователя (только админ)
func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if _, exists := users[req.Email]; !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	delete(users, req.Email)

	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("User %s deleted successfully", req.Email),
	})
}

// Вспомогательная функция для получения всех пользователей
func getAllUsers() []User {
	result := make([]User, 0, len(users))
	for _, user := range users {
		result = append(result, *user)
	}
	return result
}

func main() {
	// Публичные endpoints
	http.HandleFunc("/public", publicHandler)
	http.HandleFunc("/login", loginHandler)

	// Защищенные endpoints (требуют аутентификации)
	http.HandleFunc("/profile", authMiddleware(profileHandler))

	// Endpoints с проверкой ролей
	http.HandleFunc("/admin/users", authMiddleware(
		requireRole(RoleAdmin, adminHandler),
	))

	http.HandleFunc("/admin/delete-user", authMiddleware(
		requireRole(RoleAdmin, deleteUserHandler),
	))

	http.HandleFunc("/moderation", authMiddleware(
		requireAnyRole([]Role{RoleAdmin, RoleModerator}, moderationHandler),
	))

	fmt.Println("Server started on :8080")
	fmt.Println("\nПредустановленные пользователи:")
	fmt.Println("  Admin:     admin@example.com / admin123")
	fmt.Println("  Moderator: moderator@example.com / mod123")
	fmt.Println("  User:      user@example.com / user123")
	fmt.Println("\nПримеры запросов:")
	fmt.Println("  # Логин как admin")
	fmt.Println(`  TOKEN=$(curl -s -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"email":"admin@example.com","password":"admin123"}' | jq -r '.token')`)
	fmt.Println("\n  # Доступ к admin panel")
	fmt.Println(`  curl http://localhost:8080/admin/users -H "Authorization: Bearer $TOKEN"`)
	fmt.Println("\n  # Логин как обычный пользователь")
	fmt.Println(`  USER_TOKEN=$(curl -s -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"email":"user@example.com","password":"user123"}' | jq -r '.token')`)
	fmt.Println("\n  # Попытка доступа к admin panel (403)")
	fmt.Println(`  curl http://localhost:8080/admin/users -H "Authorization: Bearer $USER_TOKEN"`)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
