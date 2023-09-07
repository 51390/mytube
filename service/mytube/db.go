package mytube

import (
	"fmt"
    "os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDb() (db *gorm.DB, err error) {
    password := os.Getenv("POSTGRES_PASSWORD")
    dsn := fmt.Sprintf("host=localhost user=postgres password=%s dbname=mytube_service port=5432 sslmode=disable", password)
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    return
}

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(&Video{})
	HandleError(err, "Failed migrating db.")
	fmt.Println("Migrate db ok")
}
