package main

import (
	"encoding/json"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type mqttIngester struct {
	recStore          recStore
	currentStore      currentStore
	brokerAddr        string
	connectionTimeout time.Duration
	subscribeTimeout  time.Duration
}

func newMqttIngester(
	recStore recStore,
	currentStore currentStore,
	brokerAddr string,
) mqttIngester {
	return mqttIngester{
		recStore:     recStore,
		currentStore: currentStore,
		brokerAddr:   brokerAddr,
		// TODO: Make configurable
		connectionTimeout: time.Duration(5 * time.Second),
		subscribeTimeout:  time.Duration(5 * time.Second),
	}
}

func (i mqttIngester) start() {
	// TODO: Run periodically so that points may be added/removed
	recs, err := i.recStore.readRecs("mqttSubject")
	if err != nil {
		log.Fatalf("Error getting mqttSubject points: %s", err)
		return
	}

	// Configure
	mqttClient := mqtt.NewClient(mqtt.NewClientOptions().AddBroker(i.brokerAddr))

	// Connect
	connectToken := mqttClient.Connect()
	if !connectToken.WaitTimeout(i.connectionTimeout) {
		log.Fatalf("Unable to connect to %s", i.brokerAddr)
	}
	if connectToken.Error() != nil {
		log.Fatal(connectToken.Error())
	}
	log.Printf("Connected to %s", i.brokerAddr)
	defer func() {
		mqttClient.Disconnect(1)
		log.Printf("Disconnected from %s", i.brokerAddr)
	}()

	// Stores a list of subject names that have been subscribed to. We use a map because go does not have a set type.
	var subjects = map[string][]rec{}
	for _, record := range recs {
		subject, ok := record.Tags["mqttSubject"].(string)
		if !ok {
			log.Fatal("Error asserting type for mqttSubject")
		}
		if subjects[subject] != nil {
			subjects[subject] = append(subjects[subject], record)
		} else {
			subjects[subject] = []rec{record}
		}
	}

	// Subscribe to each subject
	for subject, recs := range subjects {
		subscribeToken := mqttClient.Subscribe(
			subject,
			0,
			func(c mqtt.Client, m mqtt.Message) {
				// TODO: Change messages to JSON
				// var currentItem currentInput
				var currentItem float64
				err = json.Unmarshal(m.Payload(), &currentItem)
				if err != nil {
					log.Printf("Cannot decode message JSON: %s", err)
					return
				}
				for _, rec := range recs {
					i.currentStore.setCurrent(rec.ID, currentInput{Value: &currentItem})
				}
			},
		)
		if !subscribeToken.WaitTimeout(i.subscribeTimeout) {
			log.Printf("Unable to subscribe to %s", i.brokerAddr)
		}
		if subscribeToken.Error() != nil {
			log.Print(subscribeToken.Error())
		}
		log.Printf("Subscribed to %s", subject)
	}

	defer func() {
		for subject, _ := range subjects {
			mqttClient.Unsubscribe(subject)
			log.Printf("Unsubscribed from %s", subject)
		}
	}()

	// Blocks indefinitely
	block := make(chan int)
	<-block
}
