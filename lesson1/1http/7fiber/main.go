package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New(fiber.Config{
		Immutable: true,
	})
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.Status(200).SendString(c.OriginalURL())

	})
	app.Post("/*", func(c *fiber.Ctx) error {
		return c.Status(201).SendString(c.OriginalURL())

	})
	log.SetLevel(log.LevelInfo)
	app.Use(logger.New())
	app.Listen(":80")
}
