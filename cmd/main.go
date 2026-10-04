package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/muallimmaafi/siakad-mini/config"
	"github.com/muallimmaafi/siakad-mini/database"
	"github.com/muallimmaafi/siakad-mini/internal/middleware"
	"github.com/muallimmaafi/siakad-mini/pkg/response"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("File .env tidak ditemukan, pakai environment sistem")
	}

	db := config.ConnectDB()
	database.Migrate(db)

	app := fiber.New(fiber.Config{
		ErrorHandler: response.ErrorHandler,
	})
	app.Use(recover.New())

	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "OK", nil)
	})

	// SEMENTARA: cuma buat tes middleware, nanti dihapus
	api.Get("/ping", middleware.AuthRequired(db), middleware.RequireRole("admin"), func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "pong", nil)
	})

	log.Fatal(app.Listen(":" + os.Getenv("APP_PORT")))
}