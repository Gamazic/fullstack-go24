package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	//connStr := "user=postgres dbname=mydb sslmode=disable password=password"
	connStr := "postgres://postgres:password@localhost:5432/mydb"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := fiber.New()
	app.Post("/user", func(c *fiber.Ctx) error {
		user := User{}
		if err := c.BodyParser(&user); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(&ErrorResponse{Error: err.Error()})
		}
		if user.Password == nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(&ErrorResponse{Error: "field password should not be empty"})
		}
		encrypted, err := bcrypt.GenerateFromPassword([]byte(*user.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&ErrorResponse{Error: err.Error()})
		}
		err = db.QueryRow("INSERT INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING users_id;",
			user.Username, encrypted, user.Email).Scan(&user.Id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&ErrorResponse{Error: err.Error()})
		}
		return c.Status(fiber.StatusCreated).JSON(&user)
	})
	app.Get("/user", func(c *fiber.Ctx) error {
		return nil
	})
	app.Delete("/user/:id", func(c *fiber.Ctx) error {
		// иденификация - проверка наличия предоставленной идентичности в нашей системе (например, email)
		// аутентификация - проверяет, что запрос идет от лица идентичности (например, с помощью пароля)
		// авторизация - проверяет, на что у пользователя есть права

		header := http.Header(c.GetReqHeaders())
		// "Authorization: Basic <login:password>"
		authorization := header.Get("Authorization")
		if authorization == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(&ErrorResponse{Error: "need authorization login:password"})
		}
		// "Basic <login:password>"
		token := authorization[6:]
		logPass := strings.Split(token, ":")
		if len(logPass) != 2 {
			return c.Status(fiber.StatusUnauthorized).JSON(&ErrorResponse{Error: "need authorization login:password"})
		}
		login := logPass[0]
		password := logPass[1]

		id, err := c.ParamsInt("id", 0)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(&ErrorResponse{Error: err.Error()})
		}
		if id == 0 {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(&ErrorResponse{Error: "field id should not be zero"})
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&ErrorResponse{Error: err.Error()})
		}

		var queriedUsername string
		var queriedHash string
		err = tx.QueryRow("SELECT username, password FROM users WHERE users_id = $1;", id).
			Scan(&queriedUsername, queriedHash)
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(&ErrorResponse{Error: "user not found"})
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&ErrorResponse{Error: err.Error()})
		}

		err = bcrypt.CompareHashAndPassword([]byte(queriedHash), []byte(password))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(&ErrorResponse{Error: "bad password"})
		}

		if queriedUsername != login {
			return c.Status(fiber.StatusForbidden).JSON(&ErrorResponse{Error: fmt.Sprintf("cannot delete user with id %d", id)})
		}

		_, err = tx.Exec("DELETE FROM users WHERE users_id = $1;", id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&ErrorResponse{Error: err.Error()})
		}
		err = tx.Commit()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&ErrorResponse{Error: err.Error()})
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	app.Delete("/user_fast/:id", func(c *fiber.Ctx) error {
		// иденификация - проверка наличия предоставленной идентичности в нашей системе (например, email)
		// аутентификация - проверяет, что запрос идет от лица идентичности (например, с помощью пароля)
		// авторизация - проверяет, на что у пользователя есть права

		header := http.Header(c.GetReqHeaders())
		// "Authorization: Basic <login:password>"
		authorization := header.Get("Authorization")
		if authorization == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(&ErrorResponse{Error: "need authorization login:password"})
		}
		// "Basic <login:password>"
		token := authorization[6:]
		logPass := strings.Split(token, ":")
		if len(logPass) != 2 {
			return c.Status(fiber.StatusUnauthorized).JSON(&ErrorResponse{Error: "need authorization login:password"})
		}
		login := logPass[0]
		password := logPass[1]

		id, err := c.ParamsInt("id", 0)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(&ErrorResponse{Error: err.Error()})
		}
		if id == 0 {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(&ErrorResponse{Error: "field id should not be zero"})
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&ErrorResponse{Error: err.Error()})
		}

		r, err := db.Exec("DELETE FROM users WHERE users_id = $1 AND password = $2 RETURNING ;", id, hashedPassword)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&ErrorResponse{Error: err.Error()})
		}
		numRows, err := r.RowsAffected()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&ErrorResponse{Error: err.Error()})
		}
		if numRows == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(&ErrorResponse{Error: "no user or bad password"})
		}
		return c.SendStatus(fiber.StatusNoContent)
	})
	app.Listen(":8080")
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type User struct {
	Id       *int
	Username *string
	Password *string
	Email    *string
}

func newString(s string) *string {
	return &s
}
