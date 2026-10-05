package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/muallimmaafi/siakad-mini/internal/model"
	"github.com/muallimmaafi/siakad-mini/internal/repository"
)

// BusinessError dipakai untuk error aturan bisnis (409, 422, 403, 404)
// yang membawa status code dan pesannya sendiri.
type BusinessError struct {
	Status  int
	Message string
}

func (e *BusinessError) Error() string { return e.Message }

type EnrollmentCourse struct {
	KodeMK   string `json:"kode_mk"`
	NamaMK   string `json:"nama_mk"`
	SKS      int    `json:"sks"`
	Semester int    `json:"semester"`
}

type EnrollmentResponse struct {
	ID            uint             `json:"id"`
	StudentID     uint             `json:"student_id"`
	CourseID      uint             `json:"course_id"`
	TahunAkademik string           `json:"tahun_akademik"`
	CreatedAt     time.Time        `json:"created_at"`
	Course        EnrollmentCourse `json:"course"`
	TotalSKS      int              `json:"total_sks"`
	BatasSKS      int              `json:"batas_sks"`
	SisaSKS       int              `json:"sisa_sks"`
}

type EnrollmentService struct {
	repo *repository.EnrollmentRepository
}

func NewEnrollmentService(repo *repository.EnrollmentRepository) *EnrollmentService {
	return &EnrollmentService{repo: repo}
}

func (s *EnrollmentService) Enroll(userID, courseID uint, tahun string) (*EnrollmentResponse, error) {
	var result *EnrollmentResponse

	err := s.repo.Transaction(func(tx *repository.EnrollmentRepository) error {
		// 1. Kunci baris mahasiswa (mahasiswa hanya boleh mengubah KRS miliknya sendiri)
		student, err := tx.LockStudentByUserID(userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &BusinessError{Status: http.StatusForbidden, Message: "Hanya mahasiswa aktif yang dapat mengambil mata kuliah"}
			}
			return err
		}

		// 2. Kunci baris mata kuliah (row locking untuk kuota)
		course, err := tx.LockCourse(courseID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &ValidationError{Errors: map[string][]string{
					"course_id": {"Mata kuliah tidak ditemukan"},
				}}
			}
			return err
		}

		// 3. Duplikasi -> 409
		exists, err := tx.Exists(student.ID, course.ID, tahun)
		if err != nil {
			return err
		}
		if exists {
			return &BusinessError{Status: http.StatusConflict, Message: "Mata kuliah ini sudah Anda ambil pada tahun akademik " + tahun}
		}

		// 4. Kuota penuh -> 422
		terisi, err := tx.CountByCourse(course.ID, tahun)
		if err != nil {
			return err
		}
		if int(terisi) >= course.Kuota {
			return &BusinessError{Status: http.StatusUnprocessableEntity, Message: "Kuota mata kuliah " + course.KodeMK + " sudah penuh"}
		}

		// 5. Batas SKS -> 422 (pesan menyebut sisa SKS)
		diambil, err := tx.SumSKS(student.ID, tahun)
		if err != nil {
			return err
		}
		batas := BatasSKS(student.IPKTerakhir)
		if diambil+course.SKS > batas {
			sisa := batas - diambil
			if sisa < 0 {
				sisa = 0
			}
			return &BusinessError{
				Status: http.StatusUnprocessableEntity,
				Message: fmt.Sprintf(
					"Total SKS melebihi batas. Batas SKS Anda %d, sudah diambil %d, sisa %d SKS, sedangkan mata kuliah ini %d SKS",
					batas, diambil, sisa, course.SKS,
				),
			}
		}

		// 6. Simpan
		enrollment := &model.Enrollment{
			StudentID:     student.ID,
			CourseID:      course.ID,
			TahunAkademik: tahun,
		}
		if err := tx.Create(enrollment); err != nil {
			return err
		}

		total := diambil + course.SKS
		result = &EnrollmentResponse{
			ID:            enrollment.ID,
			StudentID:     enrollment.StudentID,
			CourseID:      enrollment.CourseID,
			TahunAkademik: enrollment.TahunAkademik,
			CreatedAt:     enrollment.CreatedAt,
			Course: EnrollmentCourse{
				KodeMK:   course.KodeMK,
				NamaMK:   course.NamaMK,
				SKS:      course.SKS,
				Semester: course.Semester,
			},
			TotalSKS: total,
			BatasSKS: batas,
			SisaSKS:  batas - total,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *EnrollmentService) Cancel(userID, enrollmentID uint) error {
	enrollment, err := s.repo.FindByID(enrollmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &BusinessError{Status: http.StatusNotFound, Message: "Data KRS tidak ditemukan"}
		}
		return err
	}

	student, err := s.repo.FindStudentByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &BusinessError{Status: http.StatusForbidden, Message: "Hanya mahasiswa aktif yang dapat mengelola KRS"}
		}
		return err
	}

	// KRS milik mahasiswa lain -> 403
	if enrollment.StudentID != student.ID {
		return &BusinessError{Status: http.StatusForbidden, Message: "Anda tidak berhak membatalkan KRS milik mahasiswa lain"}
	}

	return s.repo.Delete(enrollment.ID)
}