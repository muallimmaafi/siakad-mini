package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/muallimmaafi/siakad-mini/internal/model"
	"github.com/muallimmaafi/siakad-mini/pkg/jwtutil"
	"github.com/muallimmaafi/siakad-mini/pkg/response"
)

func AuthRequired(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		parts := strings.SplitN(c.Get("Authorization"), " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return response.Error(c, fiber.StatusUnauthorized, "Token tidak ditemukan")
		}

		claims, err := jwtutil.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "Token tidak valid atau sudah kedaluwarsa")
		}

		// Mahasiswa yang sudah di-soft delete tidak boleh lagi pakai token lamanya.
		// GORM otomatis mengabaikan record yang deleted_at-nya terisi.
		if claims.Role == "mahasiswa" {
			var count int64
			if err := db.Model(&model.Student{}).Where("user_id = ?", claims.UserID).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return response.Error(c, fiber.StatusUnauthorized, "Akun tidak aktif")
			}
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}
		return response.Error(c, fiber.StatusForbidden, "Anda tidak memiliki akses ke resource ini")
	}
}