package response

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
)

type Meta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	LastPage    int   `json:"last_page"`
}

type body struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    interface{}         `json:"data,omitempty"`
	Meta    *Meta               `json:"meta,omitempty"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

func Success(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(body{Success: true, Message: message, Data: data})
}

func SuccessWithMeta(c *fiber.Ctx, message string, data interface{}, meta Meta) error {
	return c.Status(fiber.StatusOK).JSON(body{Success: true, Message: message, Data: data, Meta: &meta})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(body{Success: false, Message: message})
}

func ValidationError(c *fiber.Ctx, errs map[string][]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(body{
		Success: false,
		Message: "Validasi gagal",
		Errors:  errs,
	})
}

// ErrorHandler dipasang di fiber.Config supaya semua error tak terduga
// jadi JSON seragam dan TIDAK membocorkan detail internal (stack trace, dll).
func ErrorHandler(c *fiber.Ctx, err error) error {
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return Error(c, fe.Code, fe.Message)
	}

	log.Printf("[ERROR] %s %s: %v", c.Method(), c.Path(), err)
	return Error(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
}