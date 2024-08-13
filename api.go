package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
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
	db.AutoMigrate(&his{})

	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	// GET /his/:pointId?start=...&end=...
	// Note that start and end are in seconds since epoch (1970-01-01T00:00:00Z)
	http.HandleFunc("GET /his/{pointId}", func(w http.ResponseWriter, request *http.Request) {
		pointIdString := request.PathValue("pointId")
		pointId, err := uuid.Parse(pointIdString)
		if err != nil {
			log.Printf("Invalid UUID: %s", pointIdString)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		err = request.ParseForm()
		if err != nil {
			log.Printf("Cannot parse form: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		params := request.Form

		var sqlResult []his
		query := db.Where(&his{PointId: pointId})
		// TODO: Change start/end to ISO8601
		if params["start"] != nil {
			startStr := params["start"][0]
			start, err := strconv.ParseInt(startStr, 0, 64)
			if err != nil {
				log.Printf("Cannot parse time: %s", startStr)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			query.Where("ts >= ?", time.Unix(start, 0))
		}
		if params["end"] != nil {
			endStr := params["end"][0]
			end, err := strconv.ParseInt(endStr, 0, 64)
			if err != nil {
				log.Printf("Cannot parse time: %s", endStr)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			query.Where("ts < ?", time.Unix(end, 0))
		}
		query.Order("ts asc").Find(&sqlResult)

		var httpResult []apiHis
		for _, sqlRow := range sqlResult {
			var value *float64
			if sqlRow.Value.Valid {
				value = &sqlRow.Value.Float64
			}
			httpResult = append(httpResult, apiHis{PointId: sqlRow.PointId, Ts: sqlRow.Ts, Value: value})
		}

		httpJson, err := json.Marshal(httpResult)
		if err != nil {
			log.Printf("Cannot encode response JSON")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(httpJson)
	})

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
