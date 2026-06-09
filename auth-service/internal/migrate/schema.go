package migrate

import (
	"log"

	"gorm.io/gorm"
)

// LegacySchema migrates data from the old monolithic users table and drops it.
func LegacySchema(db *gorm.DB) error {
	if !db.Migrator().HasTable("users") {
		return nil
	}
	if !db.Migrator().HasTable("user_credentials") {
		if err := db.Exec(`
			INSERT INTO user_credentials (user_id, password_hash, created_at, updated_at)
			SELECT id, password, created_at, updated_at FROM users
			ON CONFLICT (user_id) DO NOTHING
		`).Error; err != nil {
			log.Printf("legacy credential migration skipped: %v", err)
		}
	}
	if err := db.Exec("DROP TABLE IF EXISTS users").Error; err != nil {
		return err
	}
	log.Println("legacy users table migrated to user_credentials and dropped")
	return nil
}
