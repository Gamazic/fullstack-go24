package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// Настройки OAuth2
var oauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GITHUB_CLIENT_ID"),     // Укажите ваш Client ID
	ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"), // Укажите ваш Client Secret
	RedirectURL:  "http://localhost:8080/callback",  // URL для обратного вызова
	Scopes:       []string{"read:user", "user:email"},
	Endpoint:     github.Endpoint,
}

// Хранилище токенов (для примера, в памяти)
var tokenStore = map[string]*oauth2.Token{}

func main() {
	app := fiber.New()

	//// Главная страница с кнопкой для входа
	app.Get("/", func(c *fiber.Ctx) error {
		loginURL := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOnline)

		html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Login with GitHub</title>
		</head>
		<body>
			<h1>Login with GitHub</h1>
			<a href="%s">
				<button style="padding: 10px 20px; font-size: 16px; background-color: #24292e; color: white; border: none; border-radius: 5px; cursor: pointer;">
					Login with GitHub
				</button>
			</a>
		</body>
		</html>
		`, loginURL)

		return c.Type("html").SendString(html)
	})

	// Обработчик для callback
	app.Get("/callback", func(c *fiber.Ctx) error {
		// Проверяем, есть ли код авторизации
		code := c.Query("code")
		if code == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Code not provided")
		}

		// Обмениваем код на токен
		token, err := oauthConfig.Exchange(context.Background(), code)
		if err != nil {
			log.Printf("Exchange failed: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to exchange token")
		}

		// Сохраняем токен (для примера, в памяти)
		tokenStore["user"] = token

		return c.SendString("Login successful! You can now access the /profile endpoint.")
	})

	// Эндпоинт для получения профиля
	app.Get("/profile", func(c *fiber.Ctx) error {
		// Извлекаем токен
		token := tokenStore["user"]
		if token == nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Please login first")
		}

		// Создаём клиент с токеном
		client := oauthConfig.Client(context.Background(), token)

		// Запрашиваем данные профиля
		resp, err := client.Get("https://api.github.com/user")
		if err != nil {
			log.Printf("Failed to fetch user: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch user profile")
		}
		defer resp.Body.Close()

		// Читаем ответ
		var body []byte
		resp.Body.Read(body)

		fmt.Println(string(body))

		return c.SendString("Hi!!!!! =-)")
	})

	// Запуск сервера
	log.Fatal(app.Listen(":8080"))
}
