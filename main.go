package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %s", err)
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

	// OTEL
	metricExporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName("api-server"),
				semconv.ServiceVersion("0.1.0"),
			),
		),
		metric.WithReader(metricExporter.Reader),
	)
	otel.SetMeterProvider(meterProvider) // Sets global
	go serveMetrics()

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

	host := envOrDefault("HOST", "localhost")
	port, err := strconv.ParseInt(envOrDefault("PORT", "80"), 10, 0)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Serving at http://%s:%d", host, port)
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

func serveMetrics() {
	log.Printf("Serving metrics at localhost:2112/metrics")
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2112", nil)
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
}
