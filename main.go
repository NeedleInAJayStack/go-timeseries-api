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

	// Auth setup
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	jwtSecret := os.Getenv("JWT_SECRET")

	// Adds his
	db.AutoMigrate(&his{}, &rec{})

	registerAuth(jwtSecret, username, password)

	hisController := hisController{db: db}
	http.HandleFunc("GET /his/{pointId}", hisController.getHis)
	http.HandleFunc("POST /his/{pointId}", hisController.postHis)
	http.HandleFunc("DELETE /his/{pointId}", hisController.deleteHis)

	recController := recController{db: db}
	http.HandleFunc("GET /recs", recController.getRecs)
	http.HandleFunc("POST /recs", recController.postRecs)
	http.HandleFunc("GET /recs/tag/{tag}", recController.getRecsByTag)
	http.HandleFunc("GET /recs/{id}", recController.getRec)
	http.HandleFunc("PUT /recs/{id}", recController.putRec)
	http.HandleFunc("DELETE /recs/{id}", recController.deleteRec)

	port := 8080
	log.Printf("Serving at http://localhost:%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func envOrDefault(name string, def string) string {
	value, ok := os.LookupEnv(name)
	if ok {
		return value
	} else {
		return def
	}
}
