package main

import (
	"bobbybookmarks/internal/api"
	"bobbybookmarks/internal/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"log"
	"os"
)

func main() {
	database.InitDB()
	database.EnsureSchema()
	app := fiber.New()
	app.Use(cors.New())

	api.RegisterRoutes(app)

	// Serve static files from the 'dist' directory
	app.Static("/", "./dist")

	// Catch-all route to serve index.html for SPA (React Router)
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./dist/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Printf("Server starting on port %s...", port)
	log.Fatal(app.Listen(":" + port))
}
