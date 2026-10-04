package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/muallimmaafi/siakad-mini/internal/repository"
	"github.com/muallimmaafi/siakad-mini/internal/service"
	"github.com/muallimmaafi/siakad-mini/pkg/response"
	"github.com/muallimmaafi/siakad-mini/pkg/validation"
)

type StudentHandler struct {
	svc *service.StudentService
}

func NewStudentHandler(svc *service.StudentService) *StudentHandler {
	return &StudentHandler{svc: svc}
}

type createStudentRequest struct {
	NIM         string   `json:"nim" validate:"required,len=12,numeric"`
	Nama        string   `json:"nama" validate:"required"`
	Email       string   `json:"email" validate:"required,email"`
	Prodi       string   `json:"prodi" validate:"required"`
	Angkatan    int      `json:"angkatan" validate:"required"`
	IPKTerakhir *float64 `json:"ipk_terakhir" validate:"omitempty,gte=0,lte=4"`
}

type updateStudentRequest struct {
	NIM         *string  `json:"nim"`
	Nama        string   `json:"nama" validate:"required"`
	Prodi       string   `json:"prodi" validate:"required"`
	Angkatan    int      `json:"angkatan" validate:"required"`
	IPKTerakhir *float64 `json:"ipk_terakhir" validate:"omitempty,gte=0,lte=4"`
}

func parseID(c *fiber.Ctx) (uint, bool) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

// checkAngkatan: 4 digit dan tidak lebih dari tahun berjalan.
func checkAngkatan(errs map[string][]string, angkatan int) {
	if _, sudahAda := errs["angkatan"]; sudahAda || angkatan == 0 {
		return
	}
	if angkatan < 1000 || angkatan > time.Now().Year() {
		errs["angkatan"] = []string{"Angkatan harus 4 digit dan tidak lebih dari tahun berjalan"}
	}
}

func handleServiceError(c *fiber.Ctx, err error) error {
	var ve *service.ValidationError
	switch {
	case errors.As(err, &ve):
		return response.ValidationError(c, ve.Errors)
	case errors.Is(err, service.ErrStudentNotFound):
		return response.Error(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	case errors.Is(err, service.ErrForbidden):
		return response.Error(c, fiber.StatusForbidden, "Anda tidak memiliki akses ke data mahasiswa ini")
	default:
		return err // jadi 500 lewat ErrorHandler
	}
}

// GET /students
func (h *StudentHandler) List(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	perPage := c.QueryInt("per_page", 10)
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 50 {
		perPage = 50
	}

	errs := map[string][]string{}

	angkatan := 0
	if a := c.Query("angkatan"); a != "" {
		n, err := strconv.Atoi(a)
		if err != nil {
			errs["angkatan"] = []string{"Angkatan harus berupa angka"}
		} else {
			angkatan = n
		}
	}

	sort := c.Query("sort")
	if sort != "" && sort != "nama" && sort != "-ipk_terakhir" {
		errs["sort"] = []string{"Sort hanya boleh: nama atau -ipk_terakhir"}
	}

	if len(errs) > 0 {
		return response.ValidationError(c, errs)
	}

	students, total, err := h.svc.List(repository.StudentFilter{
		Page:     page,
		PerPage:  perPage,
		Prodi:    strings.TrimSpace(c.Query("prodi")),
		Angkatan: angkatan,
		Search:   strings.TrimSpace(c.Query("search")),
		Sort:     sort,
	})
	if err != nil {
		return err
	}

	lastPage := int((total + int64(perPage) - 1) / int64(perPage))
	if lastPage < 1 {
		lastPage = 1
	}

	return response.SuccessWithMeta(c, "Data mahasiswa berhasil diambil", students, response.Meta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
	})
}

// POST /students
func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var req createStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, map[string][]string{"body": {"Format JSON tidak valid"}})
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Nama = strings.TrimSpace(req.Nama)
	req.Prodi = strings.TrimSpace(req.Prodi)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	errs := validation.Struct(req)
	if errs == nil {
		errs = map[string][]string{}
	}
	checkAngkatan(errs, req.Angkatan)
	if len(errs) > 0 {
		return response.ValidationError(c, errs)
	}

	result, err := h.svc.Create(service.CreateStudentInput{
		NIM:      req.NIM,
		Nama:     req.Nama,
		Email:    req.Email,
		Prodi:    req.Prodi,
		Angkatan: req.Angkatan,
		IPK:      req.IPKTerakhir,
	})
	if err != nil {
		return handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "Mahasiswa berhasil ditambahkan", result)
}

// GET /students/:id
func (h *StudentHandler) Show(c *fiber.Ctx) error {
	id, ok := parseID(c)
	if !ok {
		return response.Error(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("user_id").(uint)

	detail, err := h.svc.Detail(id, role, userID, strings.TrimSpace(c.Query("tahun_akademik")))
	if err != nil {
		return handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Detail mahasiswa berhasil diambil", detail)
}

// PUT /students/:id
func (h *StudentHandler) Update(c *fiber.Ctx) error {
	id, ok := parseID(c)
	if !ok {
		return response.Error(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	var req updateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ValidationError(c, map[string][]string{"body": {"Format JSON tidak valid"}})
	}

	req.Nama = strings.TrimSpace(req.Nama)
	req.Prodi = strings.TrimSpace(req.Prodi)

	errs := validation.Struct(req)
	if errs == nil {
		errs = map[string][]string{}
	}
	checkAngkatan(errs, req.Angkatan)
	if len(errs) > 0 {
		return response.ValidationError(c, errs)
	}

	result, err := h.svc.Update(id, service.UpdateStudentInput{
		NIM:      req.NIM,
		Nama:     req.Nama,
		Prodi:    req.Prodi,
		Angkatan: req.Angkatan,
		IPK:      req.IPKTerakhir,
	})
	if err != nil {
		return handleServiceError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Data mahasiswa berhasil diperbarui", result)
}

// DELETE /students/:id
func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, ok := parseID(c)
	if !ok {
		return response.Error(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	if err := h.svc.Delete(id); err != nil {
		return handleServiceError(c, err)
	}

	return response.NoContent(c)
}