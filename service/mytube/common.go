package mytube

import (
    "fmt"

    "github.com/joho/godotenv"
)

func LoadEnv() {
    err := godotenv.Load(".env")
    if err != nil {
        fmt.Printf("Failed to load .env file: %s\n", err)
    }
}

