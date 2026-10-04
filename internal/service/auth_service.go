package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/muallimmaafi/siakad-mini/internal/repository"
	"github.com/muallimmaafi/siakad-mini/pkg/jwtutil"
)

var (
	ErrInvalidCredentials = errors.New("email atau password salah")
	ErrUserNotFound       = errors.New("user tidak ditemukan")
)

type UserInfo struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type LoginResult struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int      `json:"expires_in"` // detik
	User        UserInfo `json:"user"`
}

type StudentInfo struct {
	NIM      string `json:"nim"`
	Nama     string `json:"nama"`
	Prodi    string `json:"prodi"`
	Angkatan int    `json:"angkatan"`
}

type MeResult struct {
	ID      uint         `json:"id"`
	Email   string       `json:"email"`
	Role    string       `json:"role"`
	Student *StudentInfo `json:"student,omitempty"`
}

type AuthService struct {
	users *repository.UserRepository
}

func NewAuthService(users *repository.UserRepository) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Login(email, password string) (*LoginResult, error) {
	user, err := s.users.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}

	// Mahasiswa yang sudah di-soft delete tidak boleh login
	if user.Role == "mahasiswa" && user.Student == nil {
		return nil, ErrInvalidCredentials
	}

	token, err := jwtutil.Generate(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(jwtutil.ExpiresIn().Seconds()),
		User:        UserInfo{ID: user.ID, Email: user.Email, Role: user.Role},
	}, nil
}

func (s *AuthService) Me(userID uint) (*MeResult, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	result := &MeResult{ID: user.ID, Email: user.Email, Role: user.Role}
	if user.Role == "mahasiswa" && user.Student != nil {
		result.Student = &StudentInfo{
			NIM:      user.Student.NIM,
			Nama:     user.Student.Nama,
			Prodi:    user.Student.Prodi,
			Angkatan: user.Student.Angkatan,
		}
	}
	return result, nil
}