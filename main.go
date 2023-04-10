package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/tsivinsky/array"
)

type ApiError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func NewApiError(c *fiber.Ctx, code int, err error) error {
	return c.Status(code).JSON(&ApiError{code, err.Error()})
}

type Item struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

func main() {
	app := fiber.New()

	items := []Item{}

	app.Use(cors.New())
	app.Use(recover.New())

	app.Post("/", func(c *fiber.Ctx) error {
		var item Item

		if err := c.BodyParser(&item); err != nil {
			return NewApiError(c, 400, err)
		}

		if item.Name == "" || item.Source == "" {
			return NewApiError(c, 400, errors.New("No 'item' or 'source' in body"))
		}

		items = append(items, Item{
			Name:   item.Name,
			Source: item.Source,
		})

		return c.Status(201).JSON(fiber.Map{
			"message": "item added",
		})
	})

	app.All("/", func(c *fiber.Ctx) error {
		subdomains := c.Subdomains()
		if len(subdomains) == 0 {
			return NewApiError(c, 400, errors.New("no subdomain"))
		}

		subdomain := subdomains[0]

		item := array.Find(items, func(item Item, i int) bool {
			return item.Name == subdomain
		})
		if item.Name == "" {
			return NewApiError(c, 404, errors.New("No item found"))
		}

		if err := proxy.Do(c, item.Source); err != nil {
			return NewApiError(c, 500, errors.New("can't proxy it"))
		}

		return NewApiError(c, 500, fmt.Errorf("Couldn't proxy %s", item.Source))
	})

	log.Fatal(app.Listen(":5000"))
}
