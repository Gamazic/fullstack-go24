package main

import (
	"github.com/gofiber/fiber/v2"
)

func fiberHandler(c *fiber.Ctx) error {
	return c.Status(200).SendString(c.OriginalURL())
}

func main() {
	app := fiber.New(fiber.Config{
		Immutable: true,
	})
	app.Get("/*", fiberHandler)
	app.Listen(":80")
}
