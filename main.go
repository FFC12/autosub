package main

import (
	api "autosub/api"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New(fiber.Config{
		StreamRequestBody: true,
	})

	app.Post("/upload", api.Upload)

	// Serve static files from the "static" folder
	app.Static("/", "./static")
	app.Static("/video", "./processed")

	// Define the route to serve the video upload form
	app.Get("/upload-form", func(c *fiber.Ctx) error {
		return c.SendFile("./static/index.html")
	})

	app.Listen(":3000")
}
