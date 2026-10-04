package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/muallimmaafi/siakad-mini/config"
	"github.com/muallimmaafi/siakad-mini/database"
	"github.com/muallimmaafi/siakad-mini/internal/model"
)

type studentSeed struct {
	Nama     string
	Prodi    string
	Angkatan int
	IPK      float64
}

func hash(plain string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Gagal hash password: ", err)
	}
	return string(b)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("File .env tidak ditemukan, pakai environment sistem")
	}

	db := config.ConnectDB()
	database.Migrate(db)

	seedAdmin(db)
	seedStudents(db)
	seedCourses(db)

	log.Println("Seeding selesai")
}

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Where("email = ?", "admin@siakad.test").Count(&count)
	if count > 0 {
		log.Println("Admin sudah ada, dilewati")
		return
	}

	admin := model.User{
		Email:    "admin@siakad.test",
		Password: hash("admin12345"),
		Role:     "admin",
	}
	if err := db.Create(&admin).Error; err != nil {
		log.Fatal("Gagal seed admin: ", err)
	}
	log.Println("Admin dibuat: admin@siakad.test / admin12345")
}

func seedStudents(db *gorm.DB) {
	data := []studentSeed{
		{"Rina Putri", "Sistem Informasi", 2022, 3.45},
		{"Budi Santoso", "Teknik Informatika", 2022, 3.10},
		{"Siti Aminah", "Teknik Informatika", 2021, 3.80},
		{"Ahmad Fauzi", "Sistem Informasi", 2023, 2.75},
		{"Dewi Lestari", "Teknik Informatika", 2023, 2.60},
		{"Eko Prasetyo", "Sistem Informasi", 2021, 3.25},
		{"Fitri Handayani", "Teknik Informatika", 2024, 2.30},
		{"Gilang Ramadhan", "Sistem Informasi", 2022, 2.90},
		{"Hana Safitri", "Teknik Informatika", 2023, 3.55},
		{"Indra Kurniawan", "Sistem Informasi", 2024, 2.45},
		{"Joko Widodo", "Teknik Informatika", 2021, 3.00},
		{"Kartika Sari", "Sistem Informasi", 2022, 3.70},
		{"Lukman Hakim", "Teknik Informatika", 2023, 2.85},
		{"Maya Anggraini", "Sistem Informasi", 2024, 2.20},
		{"Naufal Rizky", "Teknik Informatika", 2022, 3.35},
		{"Oktaviani Putri", "Sistem Informasi", 2021, 2.55},
		{"Putra Mahendra", "Teknik Informatika", 2023, 3.15},
		{"Qonita Azzahra", "Sistem Informasi", 2024, 3.90},
		{"Rizal Firmansyah", "Teknik Informatika", 2022, 2.00},
		{"Sekar Ayu", "Sistem Informasi", 2023, 3.40},
	}

	for i, s := range data {
		nim := fmt.Sprintf("18722100%04d", i+1)
		email := fmt.Sprintf("mhs%02d@siakad.test", i+1)

		var count int64
		db.Model(&model.User{}).Where("email = ?", email).Count(&count)
		if count > 0 {
			continue
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			user := model.User{
				Email:    email,
				Password: hash(nim), // password awal = NIM
				Role:     "mahasiswa",
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}

			student := model.Student{
				UserID:      user.ID,
				NIM:         nim,
				Nama:        s.Nama,
				Prodi:       s.Prodi,
				Angkatan:    s.Angkatan,
				IPKTerakhir: s.IPK,
			}
			return tx.Create(&student).Error
		})
		if err != nil {
			log.Fatalf("Gagal seed mahasiswa %s: %v", s.Nama, err)
		}
	}
	log.Println("Mahasiswa selesai di-seed (password awal = NIM)")
}

func seedCourses(db *gorm.DB) {
	data := []model.Course{
		{KodeMK: "TI101", NamaMK: "Algoritma dan Pemrograman", SKS: 3, Semester: 1, Kuota: 30},
		{KodeMK: "TI102", NamaMK: "Matematika Diskrit", SKS: 3, Semester: 1, Kuota: 30},
		{KodeMK: "TI201", NamaMK: "Struktur Data", SKS: 3, Semester: 2, Kuota: 25},
		{KodeMK: "TI202", NamaMK: "Basis Data", SKS: 4, Semester: 2, Kuota: 25},
		{KodeMK: "TI301", NamaMK: "Pemrograman Web", SKS: 3, Semester: 3, Kuota: 20},
		{KodeMK: "TI302", NamaMK: "Rekayasa Perangkat Lunak", SKS: 3, Semester: 3, Kuota: 20},
		{KodeMK: "TI401", NamaMK: "Pemrograman Backend Lanjut", SKS: 4, Semester: 4, Kuota: 15},
		{KodeMK: "TI402", NamaMK: "Keamanan Siber", SKS: 3, Semester: 4, Kuota: 15},
		{KodeMK: "TI501", NamaMK: "Kecerdasan Buatan", SKS: 3, Semester: 5, Kuota: 10},
		{KodeMK: "TI502", NamaMK: "Praktikum Proyek Akhir", SKS: 2, Semester: 5, Kuota: 2}, // kuota kecil buat tes "kuota penuh"
	}

	for _, c := range data {
		course := c
		if err := db.Where("kode_mk = ?", course.KodeMK).FirstOrCreate(&course).Error; err != nil {
			log.Fatalf("Gagal seed matkul %s: %v", course.KodeMK, err)
		}
	}
	log.Println("Mata kuliah selesai di-seed")
}