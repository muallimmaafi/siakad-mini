package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/muallimmaafi/siakad-mini/config"
	"github.com/muallimmaafi/siakad-mini/database"
	"github.com/muallimmaafi/siakad-mini/internal/handler"
	"github.com/muallimmaafi/siakad-mini/internal/middleware"
	"github.com/muallimmaafi/siakad-mini/internal/repository"
	"github.com/muallimmaafi/siakad-mini/internal/service"
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

	// Wiring: repository -> service -> handler
	userRepo := repository.NewUserRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	courseRepo := repository.NewCourseRepository(db)

	authSvc := service.NewAuthService(userRepo)
	studentSvc := service.NewStudentService(studentRepo)
	courseSvc := service.NewCourseService(courseRepo)

	authHandler := handler.NewAuthHandler(authSvc)
	studentHandler := handler.NewStudentHandler(studentSvc)
	courseHandler := handler.NewCourseHandler(courseSvc)

	// Rate limit login: maks. 5 kali gagal per menit per IP.
	loginLimiter := limiter.New(limiter.Config{
		Max:                    5,
		Expiration:             1 * time.Minute,
		SkipSuccessfulRequests: true,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.Error(c, fiber.StatusTooManyRequests,
				"Terlalu banyak percobaan login, coba lagi dalam 1 menit")
		},
	})

	authRequired := middleware.AuthRequired(db)
	adminOnly := middleware.RequireRole("admin")

	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "OK", nil)
	})

	// Auth
	auth := api.Group("/auth")
	auth.Post("/login", loginLimiter, authHandler.Login)
	auth.Get("/me", authRequired, authHandler.Me)

	// Students
	students := api.Group("/students", authRequired)
	students.Get("/", adminOnly, studentHandler.List)
	students.Post("/", adminOnly, studentHandler.Create)
	students.Get("/:id", middleware.RequireRole("admin", "mahasiswa"), studentHandler.Show)
	students.Put("/:id", adminOnly, studentHandler.Update)
	students.Delete("/:id", adminOnly, studentHandler.Delete)

	// Courses (semua role yang sudah login)
	api.Get("/courses", authRequired, courseHandler.List)

	log.Fatal(app.Listen(":" + os.Getenv("APP_PORT")))
}