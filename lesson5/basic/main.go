package main

import (
	"encoding/base64"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func basicAuthHandler(c *fiber.Ctx) error {
	// Получаем заголовок Authorization
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header is missing",
		})
	}

	// Проверяем, начинается ли заголовок с "Basic "
	if !strings.HasPrefix(authHeader, "Basic ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid authorization method",
		})
	}

	// Декодируем Base64 строку
	encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")
	decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Base64 encoding",
		})
	}

	// Разделяем логин и пароль
	credentials := strings.SplitN(string(decodedCredentials), ":", 2)
	if len(credentials) != 2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials format",
		})
	}

	username, password := credentials[0], credentials[1]

	if checkCredentials(username, password) {
		return c.JSON(fiber.Map{
			"message": "Login successful",
			"user":    username,
		})
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": "Invalid credentials",
	})
}

func checkCredentials(username, password string) bool {
	// Проверяем логин и пароль (статические для примера)
	return username == "admin" && password == "password123"
}

func main() {
	app := fiber.New()

	// Защищённый маршрут с Basic Authentication
	app.Get("/protected", basicAuthHandler)

	app.Listen(":8080")
}
