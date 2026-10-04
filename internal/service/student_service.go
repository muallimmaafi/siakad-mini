package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/muallimmaafi/siakad-mini/internal/model"
	"github.com/muallimmaafi/siakad-mini/internal/repository"
)

var (
	ErrStudentNotFound = errors.New("mahasiswa tidak ditemukan")
	ErrForbidden       = errors.New("akses ditolak")
)

// ValidationError membawa detail error per field untuk dijadikan response 422.
type ValidationError struct {
	Errors map[string][]string
}

func (e *ValidationError) Error() string { return "validasi gagal" }

type StudentResponse struct {
	ID          uint    `json:"id"`
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
	Email       string  `json:"email,omitempty"`
}

type CourseTaken struct {
	EnrollmentID  uint   `json:"enrollment_id"`
	CourseID      uint   `json:"course_id"`
	KodeMK        string `json:"kode_mk"`
	NamaMK        string `json:"nama_mk"`
	SKS           int    `json:"sks"`
	Semester      int    `json:"semester"`
	TahunAkademik string `json:"tahun_akademik"`
}

type StudentDetail struct {
	StudentResponse
	MataKuliah []CourseTaken `json:"mata_kuliah"`
	TotalSKS   int           `json:"total_sks"`
	BatasSKS   int           `json:"batas_sks"`
}

type CreateStudentInput struct {
	NIM      string
	Nama     string
	Email    string
	Prodi    string
	Angkatan int
	IPK      *float64
}

type UpdateStudentInput struct {
	NIM      *string // kalau dikirim, harus sama dengan NIM lama
	Nama     string
	Prodi    string
	Angkatan int
	IPK      *float64
}

type StudentService struct {
	repo *repository.StudentRepository
}

func NewStudentService(repo *repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

func toResponse(s *model.Student) StudentResponse {
	return StudentResponse{
		ID:          s.ID,
		NIM:         s.NIM,
		Nama:        s.Nama,
		Prodi:       s.Prodi,
		Angkatan:    s.Angkatan,
		IPKTerakhir: s.IPKTerakhir,
	}
}

func (s *StudentService) List(f repository.StudentFilter) ([]StudentResponse, int64, error) {
	students, total, err := s.repo.List(f)
	if err != nil {
		return nil, 0, err
	}

	out := make([]StudentResponse, 0, len(students))
	for i := range students {
		out = append(out, toResponse(&students[i]))
	}
	return out, total, nil
}

// Detail: admin boleh lihat siapa saja, mahasiswa hanya data dirinya sendiri.
// tahun (opsional) menyaring total SKS per tahun akademik.
func (s *StudentService) Detail(id uint, role string, userID uint, tahun string) (*StudentDetail, error) {
	if role == "mahasiswa" {
		own, err := s.repo.FindByUserID(userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrForbidden
			}
			return nil, err
		}
		if own.ID != id {
			return nil, ErrForbidden
		}
	}

	student, err := s.repo.FindWithEnrollments(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}

	detail := &StudentDetail{
		StudentResponse: toResponse(student),
		MataKuliah:      []CourseTaken{},
		BatasSKS:        BatasSKS(student.IPKTerakhir),
	}

	for _, e := range student.Enrollments {
		if tahun != "" && e.TahunAkademik != tahun {
			continue
		}
		if e.Course == nil {
			continue
		}
		detail.MataKuliah = append(detail.MataKuliah, CourseTaken{
			EnrollmentID:  e.ID,
			CourseID:      e.CourseID,
			KodeMK:        e.Course.KodeMK,
			NamaMK:        e.Course.NamaMK,
			SKS:           e.Course.SKS,
			Semester:      e.Course.Semester,
			TahunAkademik: e.TahunAkademik,
		})
		detail.TotalSKS += e.Course.SKS
	}

	return detail, nil
}

func (s *StudentService) Create(in CreateStudentInput) (*StudentResponse, error) {
	errs := map[string][]string{}

	if exists, err := s.repo.ExistsNIM(in.NIM); err != nil {
		return nil, err
	} else if exists {
		errs["nim"] = append(errs["nim"], "NIM sudah terdaftar")
	}
	if exists, err := s.repo.ExistsEmail(in.Email); err != nil {
		return nil, err
	} else if exists {
		errs["email"] = append(errs["email"], "Email sudah terdaftar")
	}
	if len(errs) > 0 {
		return nil, &ValidationError{Errors: errs}
	}

	// Password awal = NIM (di-hash)
	hashed, err := bcrypt.GenerateFromPassword([]byte(in.NIM), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:    in.Email,
		Password: string(hashed),
		Role:     "mahasiswa",
	}
	student := &model.Student{
		NIM:      in.NIM,
		Nama:     in.Nama,
		Prodi:    in.Prodi,
		Angkatan: in.Angkatan,
	}
	if in.IPK != nil {
		student.IPKTerakhir = *in.IPK
	}

	if err := s.repo.Create(user, student); err != nil {
		return nil, err
	}

	resp := toResponse(student)
	resp.Email = user.Email
	return &resp, nil
}

func (s *StudentService) Update(id uint, in UpdateStudentInput) (*StudentResponse, error) {
	student, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}

	if in.NIM != nil && *in.NIM != student.NIM {
		return nil, &ValidationError{Errors: map[string][]string{
			"nim": {"NIM tidak dapat diubah"},
		}}
	}

	student.Nama = in.Nama
	student.Prodi = in.Prodi
	student.Angkatan = in.Angkatan
	if in.IPK != nil {
		student.IPKTerakhir = *in.IPK
	}

	if err := s.repo.Update(student); err != nil {
		return nil, err
	}

	resp := toResponse(student)
	return &resp, nil
}

func (s *StudentService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStudentNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}