package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/muallimmaafi/siakad-mini/internal/repository"
	"github.com/muallimmaafi/siakad-mini/internal/service"
	"github.com/muallimmaafi/siakad-mini/pkg/response"
)

type CourseHandler struct {
	svc *service.CourseService
}

func NewCourseHandler(svc *service.CourseService) *CourseHandler {
	return &CourseHandler{svc: svc}
}

// GET /courses
func (h *CourseHandler) List(c *fiber.Ctx) error {
	semester := 0
	if s := c.Query("semester"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			return response.ValidationError(c, map[string][]string{
				"semester": {"Semester harus berupa angka positif"},
			})
		}
		semester = n
	}

	courses, err := h.svc.List(repository.CourseFilter{
		Semester:      semester,
		Search:        strings.TrimSpace(c.Query("search")),
		OnlyAvailable: c.Query("available") == "true",
		TahunAkademik: strings.TrimSpace(c.Query("tahun_akademik")),
	})
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, "Data mata kuliah berhasil diambil", courses)
}