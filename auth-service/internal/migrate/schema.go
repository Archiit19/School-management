package migrate

import (
	"log"

	"gorm.io/gorm"
)

// UserSchema removes school_id and role_id from users — membership lives in school-service.
func UserSchema(db *gorm.DB) error {
	if err := db.Exec("ALTER TABLE users DROP COLUMN IF EXISTS school_id").Error; err != nil {
		return err
	}
	if err := db.Exec("ALTER TABLE users DROP COLUMN IF EXISTS role_id").Error; err != nil {
		return err
	}
	log.Println("users table: school_id and role_id columns removed (if present)")
	return nil
}
