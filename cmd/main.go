package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/muallimmaafi/siakad-mini/config"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("File .env tidak ditemukan, pakai environment sistem")
	}

	db := config.ConnectDB()
	_ = db // dipakai di step berikutnya

	app := fiber.New()
	app.Use(recover.New())

	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true, "message": "OK"})
	})

	log.Fatal(app.Listen(":" + os.Getenv("APP_PORT")))
}