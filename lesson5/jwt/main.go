package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Секретный ключ для подписи JWT
var jwtSecret = []byte("your_secret_key")

// Структура для хранения пользователей (в реальном приложении используйте базу данных)
var users = map[string]string{} // username: hashedPassword

// Хэндлер регистрации
func registerHandler(c *fiber.Ctx) error {
	type RegisterRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	req := new(RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Проверяем, существует ли пользователь
	if _, exists := users[req.Username]; exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "User already exists"})
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Сохраняем пользователя
	users[req.Username] = string(hashedPassword)
	return c.JSON(fiber.Map{"message": "User registered successfully"})
}

// Хэндлер логина
func loginHandler(c *fiber.Ctx) error {
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	req := new(LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Проверяем существование пользователя
	hashedPassword, exists := users[req.Username]
	if !exists {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Сравниваем пароли
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Генерируем JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"exp":      time.Now().Add(time.Hour * 1).Unix(), // Токен действует 1 час
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(fiber.Map{"access_token": tokenString})
}

// Middleware для проверки JWT
func jwtMiddleware(c *fiber.Ctx) error {
	// Получаем токен из заголовка Authorization
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or missing Authorization header"})
	}

	tokenString := authHeader[7:]

	// Парсим и проверяем токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}

	// Извлекаем данные из токена
	claims := token.Claims.(jwt.MapClaims)
	username := claims["username"].(string)

	// Добавляем данные в контекст
	c.Locals("username", username)

	return c.Next()
}

// Пример защищённого маршрута
func protectedRouteHandler(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	return c.JSON(fiber.Map{
		"message":  "Access granted",
		"username": username,
	})
}

func main() {
	app := fiber.New()

	// Маршруты
	app.Post("/register", registerHandler)
	app.Post("/login", loginHandler)
	app.Get("/protected", jwtMiddleware, protectedRouteHandler)

	// Запуск сервера
	app.Listen(":8080")
}
