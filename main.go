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

	// Database
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		envOrDefault("DATABASE_HOST", "localhost"),
		envOrDefault("DATABASE_PORT", "5432"),
		envOrDefault("DATABASE_USERNAME", "postgres"),
		envOrDefault("DATABASE_PASSWORD", "postgres"),
		envOrDefault("DATABASE_NAME", "postgres"),
		envOrDefault("DATABASE_SSL", "disable"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	serverConfig := ServerConfig{
		username:             os.Getenv("USERNAME"),
		password:             os.Getenv("PASSWORD"),
		jwtSecret:            os.Getenv("JWT_SECRET"),
		tokenDurationSeconds: 60 * 60, // 1 hour

		db:            db,
		dbAutoMigrate: true,
	}
	server, err := NewServer(serverConfig)
	if err != nil {
		log.Fatal(err)
	}

	port := 8080
	log.Printf("Serving at http://localhost:%d", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), server)
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func envOrDefault(name string, def string) string {
	value, ok := os.LookupEnv(name)
	if ok {
		return value
	} else {
		return def
	}
}
