package yourtube

import (
    "fmt"
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
)

func InitDb() (db *gorm.DB, err error) {
    dsn := "host=localhost user=postgres password=db712bccc14d602212c928a39ba7e23d dbname=yourtube port=5432 sslmode=disable"
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    return
}

func Migrate(db *gorm.DB) {
    err := db.AutoMigrate(&video{})
    HandleError(err, "Failed migrating db.")
    fmt.Println("Migrate db ok")
}
