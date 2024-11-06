package internal

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func newString(s string) *string {
	return &s
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Api struct {
	FiberApp *fiber.App
	storage  UserStorage
}

func NewApi() *Api {
	engine := html.New("./template", ".html")
	app := fiber.New(fiber.Config{Views: engine})

	api := Api{FiberApp: app, storage: NewInMemoryStorage()}

	app.Get("/admin", api.AdminPage)
	app.Delete("/admin/userMap/:id", api.DeleteUser)
	app.Post("/admin/userMap", api.CreateUser)

	return &api
}

func (a *Api) AdminPage(c *fiber.Ctx) error {
	users := a.storage.GetAll()
	return c.Render("admin", fiber.Map{
		"Users": users,
	})
}

func (a *Api) DeleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(&ErrorResponse{Error: err.Error()})
	}
	a.storage.Delete(id)
	return c.SendStatus(fiber.StatusNoContent)
}

func (a *Api) CreateUser(c *fiber.Ctx) error {
	user := User{}
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(&ErrorResponse{Error: err.Error()})
	}
	if user.Username == nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(&ErrorResponse{Error: "field username should not be empty"})
	}
	createdUser := a.storage.CreateUser(user)
	return c.Status(fiber.StatusCreated).JSON(&createdUser)
}

type UserStorage interface {
	GetAll() []User
	Delete(id int)
	CreateUser(user User) User
}
