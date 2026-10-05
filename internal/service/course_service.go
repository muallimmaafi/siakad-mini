package service

import "github.com/muallimmaafi/siakad-mini/internal/repository"

type CourseResponse struct {
	ID        uint   `json:"id"`
	KodeMK    string `json:"kode_mk"`
	NamaMK    string `json:"nama_mk"`
	SKS       int    `json:"sks"`
	Semester  int    `json:"semester"`
	Kuota     int    `json:"kuota"`
	Terisi    int    `json:"terisi"`
	SisaKuota int    `json:"sisa_kuota"`
}

type CourseService struct {
	repo *repository.CourseRepository
}

func NewCourseService(repo *repository.CourseRepository) *CourseService {
	return &CourseService{repo: repo}
}

func (s *CourseService) List(f repository.CourseFilter) ([]CourseResponse, error) {
	rows, err := s.repo.List(f)
	if err != nil {
		return nil, err
	}

	out := make([]CourseResponse, 0, len(rows))
	for _, r := range rows {
		sisa := r.Kuota - r.Terisi
		if sisa < 0 {
			sisa = 0
		}
		out = append(out, CourseResponse{
			ID:        r.ID,
			KodeMK:    r.KodeMK,
			NamaMK:    r.NamaMK,
			SKS:       r.SKS,
			Semester:  r.Semester,
			Kuota:     r.Kuota,
			Terisi:    r.Terisi,
			SisaKuota: sisa,
		})
	}
	return out, nil
}