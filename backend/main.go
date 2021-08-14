package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github/JussMor/wsl-api/postgres"

	_ "github.com/lib/pq"
)
func mapTodo(todo postgres.Todo) interface{} {
	return struct {
		ID int64 `json:"id"`
		Name string `json:"name"`
		Completed bool `json:"completed"`
	}{
		ID: todo.ID,
		Name: todo.Name,
		Completed: todo.Completed.Bool,
	}
}

type Handlers struct {
	Repo *postgres.Repo
}

func NewHandlers(repo *postgres.Repo) *Handlers {
	return &Handlers{Repo: repo}
}

func main () {
	db, err := sql.Open("postgres", fmt.Sprintf("dbname=%s password=secret user=root sslmode=disable", "simple_bank"))
	if err != nil {
		panic(err)
	}
	repo := postgres.NewRepo(db)

	app:= fiber.New()

	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/", func (c *fiber.Ctx) error {
        return c.SendString("Hello, World!")
    })

	handlers := NewHandlers(repo)

	SetupApiV1(app, handlers)

	err = app.Listen(":3000")
	if err != nil {
		panic(err)
	}

}

func SetupApiV1(app *fiber.App, handlers *Handlers) {
	v1 := app.Group("/v1")

	SetupTodosRoutes(v1, handlers)
}

func SetupTodosRoutes(grp fiber.Router, handlers *Handlers) {
	todosRoutes := grp.Group("/todos")
	todosRoutes.Get("/", handlers.GetTodos)
	todosRoutes.Post("/",handlers.CreateTodo)
	todosRoutes.Get("/:id", handlers.GetTodo)
	todosRoutes.Delete("/:id", handlers.DeleteTodo)
	todosRoutes.Patch("/:id", handlers.UpdateTodo)

}

func (h *Handlers) GetTodos (ctx *fiber.Ctx) error {
	todos, err := h.Repo.GetAllTodos(ctx.Context())
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).Send([]byte(err.Error()))
		return err
	}
	result := make([]interface{}, len(todos))
	for i, todo := range todos {
		result[i] = mapTodo(todo)
	}
	if err := ctx.Status(fiber.StatusOK).JSON(result); err != nil {
		ctx.Status(fiber.StatusInternalServerError).Send([]byte(err.Error()))
		return err
	}
	return err
}

func (h *Handlers) CreateTodo(ctx *fiber.Ctx) error {
	type request struct {
		Name string `json:"name"`
	}
	var body request
	err := ctx.BodyParser(&body)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse json",
		})
	}

	if len(body.Name) <= 2{
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name not long enough",
		})
	}
	todo, err := h.Repo.CreateTodo(ctx.Context(), body.Name)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).Send([]byte(err.Error()))	
	}
	if err := ctx.Status(fiber.StatusCreated).JSON(mapTodo(todo)); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).Send([]byte(err.Error()))
	}
 	return  err
}

func (h *Handlers) GetTodo(ctx *fiber.Ctx) error { 

	paramsId := ctx.Params("id")
	id, err := strconv.Atoi(paramsId)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse id",
		})
	}
	todo, err := h.Repo.GetTodoById(ctx.Context(), int64(id))

	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No existe ningun todo con ese id",
		})
	}
	if err := ctx.Status(fiber.StatusOK).JSON(mapTodo(todo)); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).Send([]byte(err.Error()))
	}
	return err
}

func (h *Handlers) DeleteTodo(ctx *fiber.Ctx) error { 
	paramsId := ctx.Params("id")
	id, err := strconv.Atoi(paramsId)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse id",
		})
	}

	_, err = h.Repo.GetTodoById(ctx.Context(), int64(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map {
			"error": "No ha sido encontrado",
		})
		
	}

	err = h.Repo.DeleteTodoById(ctx.Context(), int64(id))
	if err != nil {
		return ctx.SendStatus(fiber.StatusNotFound)
		
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (h *Handlers) UpdateTodo(ctx *fiber.Ctx) error {
	type request struct {
		Name      *string `json:"name"`
		Completed *bool   `json:"completed"`
	}
	paramsId := ctx.Params("id")

	id, err := strconv.Atoi(paramsId)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse id",
		})
	}
	var body request
	err = ctx.BodyParser(&body)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse body",
		})
		
	}

	todo, err := h.Repo.GetTodoById(ctx.Context(), int64(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error" : "no ha podido ser encontrado",
		})
		
	}
	if body.Name != nil {
		todo.Name = *body.Name
	}
	if body.Completed != nil {
		todo.Completed = sql.NullBool{
			Bool: *body.Completed,
			Valid: true,
		}
	}

	todo, err = h.Repo.UpdateTodo(ctx.Context(), postgres.UpdateTodoParams{
		ID:        int64(id),
		Name:      todo.Name,
		Completed: todo.Completed,
	})

	if err != nil {
		return ctx.SendStatus(fiber.StatusNotFound)
		
	}

	if err := ctx.Status(fiber.StatusOK).JSON(mapTodo(todo)); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).Send([]byte(err.Error()))
		
	}
	return err
}