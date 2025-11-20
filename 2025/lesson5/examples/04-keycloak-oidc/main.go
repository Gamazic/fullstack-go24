package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Конфигурация (в продакшене - из переменных окружения)
var (
	clientID     = "your-google-client-id"
	clientSecret = "your-google-client-secret"
	redirectURL  = "http://localhost:8080/callback"
	googleIssuer = "https://accounts.google.com"
)

// Хранилище состояний (для защиты от CSRF)
var states = make(map[string]bool)
var statesMu sync.RWMutex

// Хранилище сессий (упрощенное)
var sessions = make(map[string]map[string]interface{})
var sessionsMu sync.RWMutex

// Генерация случайного state
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// Главная страница
func homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>Google OIDC Example</title>
		<style>
			body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
			.btn { padding: 10px 20px; background: #4285f4; color: white; text-decoration: none; border-radius: 5px; }
			.info { background: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0; }
			code { background: #e9ecef; padding: 2px 6px; border-radius: 3px; }
		</style>
	</head>
	<body>
		<h1>Google OIDC Authentication Example</h1>
		<div class="info">
			<p>Это пример интеграции с Google через OpenID Connect (OIDC).</p>
			<p>Для работы примера нужны Google OAuth 2.0 credentials.</p>
		</div>
		<a href="/login" class="btn">Войти через Google</a>
		<div class="info">
			<h3>Как получить Google OAuth credentials:</h3>
			<ol>
				<li>Перейдите в <a href="https://console.cloud.google.com/" target="_blank">Google Cloud Console</a></li>
				<li>Создайте новый проект или выберите существующий</li>
				<li>Включите "Google+ API"</li>
				<li>Перейдите в "Credentials" → "Create Credentials" → "OAuth client ID"</li>
				<li>Application type: <code>Web application</code></li>
				<li>Authorized redirect URIs: <code>http://localhost:8080/callback</code></li>
				<li>Скопируйте Client ID и Client Secret</li>
				<li>Обновите <code>clientID</code> и <code>clientSecret</code> в <code>main.go</code></li>
			</ol>
		</div>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// Обработчик логина (редирект на Google)
func loginHandler(provider *oidc.Provider, oauth2Config oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Генерируем случайный state для защиты от CSRF
		state := generateState()

		// Сохраняем state
		statesMu.Lock()
		states[state] = true
		statesMu.Unlock()

		// Перенаправляем на Google
		url := oauth2Config.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusFound)
	}
}

// Обработчик callback от Google
func callbackHandler(oauth2Config oauth2.Config, verifier *oidc.IDTokenVerifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем state (защита от CSRF)
		state := r.URL.Query().Get("state")
		statesMu.RLock()
		validState := states[state]
		statesMu.RUnlock()

		if !validState {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			return
		}

		// Удаляем использованный state
		statesMu.Lock()
		delete(states, state)
		statesMu.Unlock()

		// Получаем authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No code in callback", http.StatusBadRequest)
			return
		}

		// Обмениваем code на токены
		oauth2Token, err := oauth2Config.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Извлекаем ID токен
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "No id_token in token response", http.StatusInternalServerError)
			return
		}

		// Верифицируем ID токен
		idToken, err := verifier.Verify(r.Context(), rawIDToken)
		if err != nil {
			http.Error(w, "Failed to verify ID token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Извлекаем claims (данные пользователя)
		var claims struct {
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
			Name          string `json:"name"`
			Picture       string `json:"picture"`
			Sub           string `json:"sub"` // User ID
		}
		if err := idToken.Claims(&claims); err != nil {
			http.Error(w, "Failed to parse claims: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Создаем сессию
		sessionID := generateState()
		sessionsMu.Lock()
		sessions[sessionID] = map[string]interface{}{
			"user_id": claims.Sub,
			"email":   claims.Email,
			"name":    claims.Name,
		}
		sessionsMu.Unlock()

		// Устанавливаем cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   3600 * 24,
		})

		// Перенаправляем на профиль
		http.Redirect(w, r, "/profile", http.StatusFound)
	}
}

// Обработчик профиля (защищенный)
func profileHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	sessionsMu.RLock()
	session, exists := sessions[cookie.Value]
	sessionsMu.RUnlock()

	if !exists {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>Profile</title>
		<style>
			body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
			.profile { background: #f8f9fa; padding: 20px; border-radius: 5px; }
			.btn { padding: 10px 20px; background: #dc3545; color: white; text-decoration: none; border-radius: 5px; display: inline-block; margin-top: 20px; }
		</style>
	</head>
	<body>
		<h1>Профиль пользователя</h1>
		<div class="profile">
			<p><strong>User ID:</strong> %s</p>
			<p><strong>Email:</strong> %s</p>
			<p><strong>Name:</strong> %s</p>
		</div>
		<a href="/logout" class="btn">Выйти</a>
	</body>
	</html>
	`, session["user_id"], session["email"], session["name"])

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// Обработчик выхода
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionsMu.Lock()
		delete(sessions, cookie.Value)
		sessionsMu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

// API endpoint для получения информации о пользователе
func apiProfileHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionsMu.RLock()
	session, exists := sessions[cookie.Value]
	sessionsMu.RUnlock()

	if !exists {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(session)
}

func main() {
	ctx := context.Background()

	// Настройка OIDC provider (Google)
	provider, err := oidc.NewProvider(ctx, googleIssuer)
	if err != nil {
		log.Fatalf("Failed to initialize OIDC provider: %v", err)
	}

	// OAuth2 конфигурация
	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// Verifier для проверки ID токенов
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	// Роуты
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler(provider, oauth2Config))
	http.HandleFunc("/callback", callbackHandler(oauth2Config, verifier))
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/api/profile", apiProfileHandler)

	fmt.Println("Server started on :8080")
	fmt.Println("\nБраузер: http://localhost:8080")
	fmt.Println("\nДля работы нужны Google OAuth 2.0 credentials")
	fmt.Println("См. README.md для инструкций по настройке")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
