package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	// Database setup
	dbHost := envOrDefault("DATABASE_HOST", "localhost")
	dbPort := envOrDefault("DATABASE_PORT", "5432")
	dbUser := envOrDefault("DATABASE_USERNAME", "postgres")
	dbPassword := envOrDefault("DATABASE_PASSWORD", "postgres")
	dbName := envOrDefault("DATABASE_NAME", "postgres")
	dbSsl := envOrDefault("DATABASE_SSL", "disable")
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbSsl,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Adds his
	db.AutoMigrate(&his{}, &rec{})

	registerHis(db)
	registerRecs(db)

	port := 8080
	log.Printf("Serving at http://localhost:%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func envOrDefault(name string, def string) string {
	value, ok := os.LookupEnv("DATABASE_HOST")
	if ok {
		return value
	} else {
		return def
	}
}
