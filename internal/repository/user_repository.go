package repository

import (
	"gorm.io/gorm"

	"github.com/muallimmaafi/siakad-mini/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Preload("Student") otomatis mengabaikan student yang sudah di-soft delete,
// jadi user.Student akan nil untuk mahasiswa yang sudah dihapus.
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Student").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Student").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}