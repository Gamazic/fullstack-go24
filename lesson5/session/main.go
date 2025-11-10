package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// Создаём хранилище сессий
var store = session.New()

func loginHandler(c *fiber.Ctx) error {
	// Получаем данные пользователя
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Пример проверки логина и пароля
	if req.Username == "user123" && req.Password == "mypassword" {
		// Создаём новую сессию
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create session"})
		}

		// Сохраняем информацию о пользователе в сессии
		sess.Set("username", req.Username)
		if err := sess.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save session"})
		}

		return c.JSON(fiber.Map{"message": "Login successful"})
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
}

func protectedHandler(c *fiber.Ctx) error {
	// Получаем сессию
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get session"})
	}

	// Проверяем, авторизован ли пользователь
	username := sess.Get("username")
	if username == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	return c.JSON(fiber.Map{"message": "Welcome " + username.(string)})
}

func logoutHandler(c *fiber.Ctx) error {
	// Удаляем сессию
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get session"})
	}
	sess.Destroy()
	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}

func main() {
	app := fiber.New()

	app.Post("/login", loginHandler)
	app.Get("/protected", protectedHandler)
	app.Post("/logout", logoutHandler)

	app.Listen(":8080")
}
