package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	//app.Use(cors.New(cors.Config{
	//	AllowOrigins: "*",
	//	AllowMethods: "GET",
	//	AllowHeaders: "Content-Type",
	//}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/status", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString("1.0.0")
	})

	err := app.Listen(":3000")
	if err != nil {
		fmt.Println(err)
		return
	}
}
