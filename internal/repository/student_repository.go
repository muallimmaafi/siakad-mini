package repository

import (
	"gorm.io/gorm"

	"github.com/muallimmaafi/siakad-mini/internal/model"
)

type StudentFilter struct {
	Page     int
	PerPage  int
	Prodi    string
	Angkatan int
	Search   string
	Sort     string
}

type StudentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

// Semua query di sini otomatis mengabaikan mahasiswa yang sudah di-soft delete.
func (r *StudentRepository) List(f StudentFilter) ([]model.Student, int64, error) {
	q := r.db.Model(&model.Student{})

	if f.Prodi != "" {
		q = q.Where("prodi ILIKE ?", f.Prodi)
	}
	if f.Angkatan > 0 {
		q = q.Where("angkatan = ?", f.Angkatan)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("(nim ILIKE ? OR nama ILIKE ?)", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "id ASC"
	switch f.Sort {
	case "nama":
		order = "nama ASC, id ASC"
	case "-ipk_terakhir":
		order = "ipk_terakhir DESC, id ASC"
	}

	var students []model.Student
	err := q.Order(order).
		Limit(f.PerPage).
		Offset((f.Page - 1) * f.PerPage).
		Find(&students).Error
	return students, total, err
}

func (r *StudentRepository) FindByID(id uint) (*model.Student, error) {
	var s model.Student
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepository) FindByUserID(userID uint) (*model.Student, error) {
	var s model.Student
	if err := r.db.Where("user_id = ?", userID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepository) FindWithEnrollments(id uint) (*model.Student, error) {
	var s model.Student
	err := r.db.Preload("Enrollments.Course").First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Unscoped: NIM milik mahasiswa yang sudah di-soft delete tetap dianggap terpakai
// (karena unique index di database masih menghitungnya).
func (r *StudentRepository) ExistsNIM(nim string) (bool, error) {
	var count int64
	err := r.db.Unscoped().Model(&model.Student{}).Where("nim = ?", nim).Count(&count).Error
	return count > 0, err
}

func (r *StudentRepository) ExistsEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// Create membuat user + student dalam SATU transaction.
// Kalau salah satu gagal, dua-duanya dibatalkan.
func (r *StudentRepository) Create(user *model.User, student *model.Student) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		student.UserID = user.ID
		return tx.Create(student).Error
	})
}

func (r *StudentRepository) Update(student *model.Student) error {
	return r.db.Save(student).Error
}

// Delete = soft delete (GORM mengisi kolom deleted_at).
func (r *StudentRepository) Delete(id uint) error {
	return r.db.Delete(&model.Student{}, id).Error
}