package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/muallimmaafi/siakad-mini/internal/model"
)

type EnrollmentRepository struct {
	db *gorm.DB
}

func NewEnrollmentRepository(db *gorm.DB) *EnrollmentRepository {
	return &EnrollmentRepository{db: db}
}

// Transaction menjalankan fn di dalam satu transaction.
// Repository "tx" yang diberikan ke fn memakai koneksi transaction yang sama.
// Kalau fn mengembalikan error, semua perubahan di-rollback.
func (r *EnrollmentRepository) Transaction(fn func(tx *EnrollmentRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(&EnrollmentRepository{db: tx})
	})
}

// SELECT ... FOR UPDATE: baris mahasiswa dikunci sampai transaction selesai.
func (r *EnrollmentRepository) LockStudentByUserID(userID uint) (*model.Student, error) {
	var s model.Student
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// SELECT ... FOR UPDATE: baris mata kuliah dikunci (row locking untuk kuota).
func (r *EnrollmentRepository) LockCourse(id uint) (*model.Course, error) {
	var c model.Course
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *EnrollmentRepository) Exists(studentID, courseID uint, tahun string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Enrollment{}).
		Where("student_id = ? AND course_id = ? AND tahun_akademik = ?", studentID, courseID, tahun).
		Count(&count).Error
	return count > 0, err
}

func (r *EnrollmentRepository) CountByCourse(courseID uint, tahun string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Enrollment{}).
		Where("course_id = ? AND tahun_akademik = ?", courseID, tahun).
		Count(&count).Error
	return count, err
}

// SumSKS = total SKS yang sudah diambil mahasiswa pada tahun akademik tertentu.
func (r *EnrollmentRepository) SumSKS(studentID uint, tahun string) (int, error) {
	var total int
	err := r.db.Table("enrollments").
		Select("COALESCE(SUM(courses.sks), 0)").
		Joins("JOIN courses ON courses.id = enrollments.course_id").
		Where("enrollments.student_id = ? AND enrollments.tahun_akademik = ?", studentID, tahun).
		Scan(&total).Error
	return total, err
}

func (r *EnrollmentRepository) Create(e *model.Enrollment) error {
	return r.db.Create(e).Error
}

func (r *EnrollmentRepository) FindByID(id uint) (*model.Enrollment, error) {
	var e model.Enrollment
	if err := r.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EnrollmentRepository) FindStudentByUserID(userID uint) (*model.Student, error) {
	var s model.Student
	if err := r.db.Where("user_id = ?", userID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *EnrollmentRepository) Delete(id uint) error {
	return r.db.Delete(&model.Enrollment{}, id).Error
}