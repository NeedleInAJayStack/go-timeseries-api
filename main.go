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

	// Auth setup
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenDurationSeconds := 60 * 60 // 1 hour

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

	db.AutoMigrate(&his{}, &rec{})

	// Register paths
	server := http.NewServeMux()
	authController := authController{
		jwtSecret:            jwtSecret,
		tokenDurationSeconds: tokenDurationSeconds,
		username:             username,
		password:             password,
	}
	server.HandleFunc("GET /auth/token", authController.getAuthToken)

	tokenAuth := http.NewServeMux()

	hisController := hisController{db: db}
	tokenAuth.HandleFunc("GET /his/{pointId}", hisController.getHis)
	tokenAuth.HandleFunc("POST /his/{pointId}", hisController.postHis)
	tokenAuth.HandleFunc("DELETE /his/{pointId}", hisController.deleteHis)
	server.Handle("/his/", tokenAuthMiddleware(jwtSecret, tokenAuth))

	recController := recController{db: db}
	tokenAuth.HandleFunc("GET /recs", recController.getRecs)
	tokenAuth.HandleFunc("POST /recs", recController.postRecs)
	tokenAuth.HandleFunc("GET /recs/tag/{tag}", recController.getRecsByTag)
	tokenAuth.HandleFunc("GET /recs/{id}", recController.getRec)
	tokenAuth.HandleFunc("PUT /recs/{id}", recController.putRec)
	tokenAuth.HandleFunc("DELETE /recs/{id}", recController.deleteRec)
	server.Handle("/recs/", tokenAuthMiddleware(jwtSecret, tokenAuth))

	port := 8080
	log.Printf("Serving at http://localhost:%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), server))
}

func envOrDefault(name string, def string) string {
	value, ok := os.LookupEnv(name)
	if ok {
		return value
	} else {
		return def
	}
}
