package mytube

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDb() (db *gorm.DB, err error) {
	password := os.Getenv("POSTGRES_PASSWORD")
	user := os.Getenv("POSTGRES_USER")
	host := os.Getenv("POSTGRES_HOST")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=mytube_service port=5432 sslmode=disable", host, user, password)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return
}

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(&Video{})
	HandleError(err, "Failed migrating db.")
	fmt.Println("Migrate db ok")
}
