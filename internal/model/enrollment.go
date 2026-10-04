package model

import "time"

type Enrollment struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	StudentID     uint      `gorm:"not null;uniqueIndex:idx_enroll_unique" json:"student_id"`
	CourseID      uint      `gorm:"not null;uniqueIndex:idx_enroll_unique" json:"course_id"`
	TahunAkademik string    `gorm:"size:30;not null;uniqueIndex:idx_enroll_unique" json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`

	Course *Course `gorm:"foreignKey:CourseID" json:"course,omitempty"`
}