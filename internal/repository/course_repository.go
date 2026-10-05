package repository

import "gorm.io/gorm"

type CourseFilter struct {
	Semester      int
	Search        string
	OnlyAvailable bool
	TahunAkademik string
}

// CourseRow adalah hasil query matkul + jumlah peserta (terisi).
type CourseRow struct {
	ID       uint   `gorm:"column:id"`
	KodeMK   string `gorm:"column:kode_mk"`
	NamaMK   string `gorm:"column:nama_mk"`
	SKS      int    `gorm:"column:sks"`
	Semester int    `gorm:"column:semester"`
	Kuota    int    `gorm:"column:kuota"`
	Terisi   int    `gorm:"column:terisi"`
}

type CourseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

func (r *CourseRepository) List(f CourseFilter) ([]CourseRow, error) {
	// Subquery: jumlah enrollment per matkul (opsional per tahun akademik)
	sub := r.db.Table("enrollments").
		Select("course_id, COUNT(*) AS terisi").
		Group("course_id")
	if f.TahunAkademik != "" {
		sub = sub.Where("tahun_akademik = ?", f.TahunAkademik)
	}

	q := r.db.Table("courses").
		Select("courses.id, courses.kode_mk, courses.nama_mk, courses.sks, courses.semester, courses.kuota, COALESCE(e.terisi, 0) AS terisi").
		Joins("LEFT JOIN (?) AS e ON e.course_id = courses.id", sub)

	if f.Semester > 0 {
		q = q.Where("courses.semester = ?", f.Semester)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("(courses.kode_mk ILIKE ? OR courses.nama_mk ILIKE ?)", like, like)
	}
	if f.OnlyAvailable {
		q = q.Where("courses.kuota > COALESCE(e.terisi, 0)")
	}

	var rows []CourseRow
	err := q.Order("courses.semester ASC, courses.kode_mk ASC").Scan(&rows).Error
	return rows, err
}