package orchestrator

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// Server encapsulates the Fiber app for the TormentNexus Orchestrator
type Server struct {
	App *fiber.App
}

// NewServer initializes the Fiber application and registers the routes
func NewServer() *Server {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
		AppName:               "TormentNexus / TormentNexus Orchestrator",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Register Routes
	setupRoutes(app)

	return &Server{App: app}
}

// Start launches the web server on the specified port
func (s *Server) Start(port string) error {
	log.Printf("[Core] Starting Unified API Server on port %s", port)
	return s.App.Listen(":" + port)
}

func setupRoutes(app *fiber.App) {
	// Root API Group
	api := app.Group("/api")

	// Health Endpoint (mirrors TS version)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"name":      "@tormentnexus/core-go",
			"mcpReady":  true, // To be wired to actual MCP status
			"timestamp": "now",
		})
	})

	// Stub for tRPC equivalent or REST bridges
	api.Get("/v1/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "tormentnexus_active"})
	})
}
