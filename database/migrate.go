package database

import (
	"log"

	"gorm.io/gorm"

	"github.com/muallimmaafi/siakad-mini/internal/model"
)

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.User{},
		&model.Student{},
		&model.Course{},
		&model.Enrollment{},
	)
	if err != nil {
		log.Fatal("Migrasi gagal: ", err)
	}
	log.Println("Migrasi selesai")
}