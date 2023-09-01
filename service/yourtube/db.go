package yourtube

import (
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
)

func InitDb() (db *gorm.DB, err error) {
    dsn := "host=localhost user=postgres password=db712bccc14d602212c928a39ba7e23d dbname=gorm port=5432 sslmode=disable"
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    return
}
