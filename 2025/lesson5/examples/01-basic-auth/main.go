package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Middleware для Basic Auth
func basicAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем заголовок Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Проверяем формат "Basic <credentials>"
		if !strings.HasPrefix(authHeader, "Basic ") {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		// Декодируем base64
		encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")
		decodedBytes, err := base64.StdEncoding.DecodeString(encodedCredentials)
		if err != nil {
			http.Error(w, "Invalid encoding", http.StatusUnauthorized)
			return
		}

		// Разделяем login:password
		credentials := string(decodedBytes)
		parts := strings.SplitN(credentials, ":", 2)
		if len(parts) != 2 {
			http.Error(w, "Invalid credentials format", http.StatusUnauthorized)
			return
		}

		username := parts[0]
		password := parts[1]

		// Проверяем credentials (в реальности - из БД)
		if username != "admin" || password != "secret" {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// Сохраняем username в контекст
		ctx := context.WithValue(r.Context(), "username", username)
		next(w, r.WithContext(ctx))
	}
}

// Защищенный endpoint
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	w.Write([]byte(fmt.Sprintf("Hello, %s! You have access to protected resource.", username)))
}

// Публичный endpoint
func publicHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is a public endpoint. No authentication required."))
}

func main() {
	// Публичный endpoint
	http.HandleFunc("/public", publicHandler)

	// Защищенный endpoint с Basic Auth
	http.HandleFunc("/protected", basicAuthMiddleware(protectedHandler))

	fmt.Println("Server started on :8080")
	fmt.Println("Try:")
	fmt.Println("  curl http://localhost:8080/public")
	fmt.Println("  curl -u admin:secret http://localhost:8080/protected")
	fmt.Println("  curl -u admin:wrong http://localhost:8080/protected")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
