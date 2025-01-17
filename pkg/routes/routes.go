package routes

import (
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/memnix/memnixrest/app/controllers"
	_ "github.com/memnix/memnixrest/docs" // Side effect import
	"time"

	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)
import "github.com/bytedance/sonic"

func New() *fiber.App {
	// Create new app
	app := fiber.New(
		fiber.Config{
			JSONEncoder: sonic.Marshal,
			JSONDecoder: sonic.Unmarshal,
		})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost, *",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // 2
	}))

	app.Use(cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Query("refresh") == "true" || c.Path() == "/v1/user" || c.Path() == "/v1/login" || c.Path() == "/v1/register" || c.Path() == "/v1/logout"
		},
		Expiration:   2 * time.Minute,
		CacheControl: true,
	}))

	app.Get("/swagger/*", swagger.HandlerDefault) // default

	// Api group
	v1 := app.Group("/v1")

	v1.Get("/", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusForbidden, "This is not a valid route") // Custom error
	})

	// Auth
	v1.Post("/register", controllers.Register)
	v1.Post("/login", controllers.Login)
	v1.Get("/user", controllers.User)
	v1.Post("/logout", controllers.Logout)

	v1.Get("/", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusForbidden, "This is not a valid route") // Custom error
	})

	// Register routes
	registerUserRoutes(v1) // /v1/users/
	registerDeckRoutes(v1) // /v1/decks/
	registerCardRoutes(v1) // /v1/cards/

	return app
}
