package handler

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/muallimmaafi/siakad-mini/internal/service"
	"github.com/muallimmaafi/siakad-mini/pkg/response"
	"github.com/muallimmaafi/siakad-mini/pkg/validation"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, map[string][]string{
			"body": {"Format JSON tidak valid"},
		})
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if errs := validation.Struct(req); errs != nil {
		return response.ValidationError(c, errs)
	}

	result, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return response.Error(c, fiber.StatusUnauthorized, "Email atau password salah")
		}
		return err // jadi 500 lewat ErrorHandler
	}

	return response.Success(c, fiber.StatusOK, "Login berhasil", result)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(uint)

	result, err := h.svc.Me(userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return response.Error(c, fiber.StatusUnauthorized, "User tidak ditemukan")
		}
		return err
	}

	return response.Success(c, fiber.StatusOK, "Data user berhasil diambil", result)
}