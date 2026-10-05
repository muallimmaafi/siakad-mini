package handler

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/muallimmaafi/siakad-mini/internal/service"
	"github.com/muallimmaafi/siakad-mini/pkg/response"
	"github.com/muallimmaafi/siakad-mini/pkg/validation"
)

type EnrollmentHandler struct {
	svc *service.EnrollmentService
}

func NewEnrollmentHandler(svc *service.EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{svc: svc}
}

type createEnrollmentRequest struct {
	CourseID      uint   `json:"course_id" validate:"required"`
	TahunAkademik string `json:"tahun_akademik" validate:"required"`
}

var tahunAkademikRe = regexp.MustCompile(`^(\d{4})/(\d{4})-(Ganjil|Genap)$`)

// Format: 2026/2027-Ganjil (tahun kedua harus = tahun pertama + 1)
func validTahunAkademik(s string) bool {
	m := tahunAkademikRe.FindStringSubmatch(s)
	if m == nil {
		return false
	}
	a, _ := strconv.Atoi(m[1])
	b, _ := strconv.Atoi(m[2])
	return b == a+1
}

func handleEnrollmentError(c *fiber.Ctx, err error) error {
	var be *service.BusinessError
	var ve *service.ValidationError
	switch {
	case errors.As(err, &be):
		return response.Error(c, be.Status, be.Message)
	case errors.As(err, &ve):
		return response.ValidationError(c, ve.Errors)
	default:
		return err // jadi 500 lewat ErrorHandler
	}
}

// POST /enrollments
func (h *EnrollmentHandler) Create(c *fiber.Ctx) error {
	var req createEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, map[string][]string{"body": {"Format JSON tidak valid"}})
	}

	req.TahunAkademik = strings.TrimSpace(req.TahunAkademik)

	errs := validation.Struct(req)
	if errs == nil {
		errs = map[string][]string{}
	}
	if _, sudahAda := errs["tahun_akademik"]; !sudahAda && !validTahunAkademik(req.TahunAkademik) {
		errs["tahun_akademik"] = []string{"Format tahun_akademik harus seperti 2026/2027-Ganjil"}
	}
	if len(errs) > 0 {
		return response.ValidationError(c, errs)
	}

	userID, _ := c.Locals("user_id").(uint)

	result, err := h.svc.Enroll(userID, req.CourseID, req.TahunAkademik)
	if err != nil {
		return handleEnrollmentError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "Mata kuliah berhasil ditambahkan ke KRS", result)
}

// DELETE /enrollments/:id
func (h *EnrollmentHandler) Delete(c *fiber.Ctx) error {
	id, ok := parseID(c)
	if !ok {
		return response.Error(c, fiber.StatusNotFound, "Data KRS tidak ditemukan")
	}

	userID, _ := c.Locals("user_id").(uint)

	if err := h.svc.Cancel(userID, id); err != nil {
		return handleEnrollmentError(c, err)
	}

	return response.NoContent(c)
}