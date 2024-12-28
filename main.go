package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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
	err = db.AutoMigrate(&gormHis{}, &gormRec{})
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
				semconv.ServiceName("timeseries-api"),
				semconv.ServiceVersion("0.1.0"),
			),
		),
		metric.WithReader(metricExporter.Reader),
	)
	otel.SetMeterProvider(meterProvider) // Sets global
	go serveMetrics()

	// Stores
	authenticator := singleUserAuthenticator{
		username: os.Getenv("USERNAME"),
		password: os.Getenv("PASSWORD"),
	}
	historyStore := newGormHistoryStore(db)
	recStore := newGormRecStore(db)
	currentStore := newInMemoryCurrentStore()

	serverConfig := ServerConfig{
		authenticator:        authenticator,
		apiKey:               os.Getenv("API_KEY"),
		jwtSecret:            os.Getenv("JWT_SECRET"),
		tokenDurationSeconds: 60 * 60, // 1 hour

		historyStore: historyStore,
		recStore:     recStore,
		currentStore: currentStore,
	}

	// Start MQTT
	mqttAddress := os.Getenv("MQTT_ADDRESS")
	mqttUsername := os.Getenv("MQTT_USERNAME")
	mqttPassword := os.Getenv("MQTT_PASSWORD")
	mqttConnectionTimeout := time.Duration(5 * time.Second)

	options := mqtt.NewClientOptions()
	options.AddBroker(mqttAddress)
	options.SetUsername(mqttUsername)
	options.SetPassword(mqttPassword)
	mqttClient := mqtt.NewClient(options)
	connectToken := mqttClient.Connect()
	if !connectToken.WaitTimeout(mqttConnectionTimeout) {
		log.Fatalf("MQTT timeout connecting to %s", mqttAddress)
	}
	if connectToken.Error() != nil {
		log.Fatal(connectToken.Error())
	}
	log.Printf("MQTT connected to %s", mqttAddress)
	valueEmitter := newMQTTValueEmitter(mqttClient)

	// Setup ingester
	ingester := newIngester(
		currentStore,
		&valueEmitter,
	)
	recs, err := recStore.readRecs("mqttSubject")
	if err != nil {
		log.Fatalf("error getting mqttSubject points: %s", err)
	}
	ingester.refreshSubscriptions(recs)

	defer func() {
		ingester.refreshSubscriptions([]rec{})
		mqttClient.Disconnect(1)
		log.Printf("Disconnected from %s", mqttAddress)
	}()

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
