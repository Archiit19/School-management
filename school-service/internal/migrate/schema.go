package migrate

import (
	"log"

	"gorm.io/gorm"
)

func DropRoleIDFromMemberships(db *gorm.DB) error {
	if err := db.Exec("ALTER TABLE user_schools DROP COLUMN IF EXISTS role_id").Error; err != nil {
		return err
	}
	log.Println("user_schools: role_id column removed (if present)")
	return nil
}
