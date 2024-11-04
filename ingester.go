package main

import (
	"encoding/json"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type mqttIngester struct {
	recStore     recStore
	currentStore currentStore
	brokerAddr   string
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
	// TODO: Make configurable
	timeout := time.Duration(5 * time.Second)
	mqttClient := mqtt.NewClient(mqtt.NewClientOptions().AddBroker(i.brokerAddr))

	connectToken := mqttClient.Connect()
	if !connectToken.WaitTimeout(timeout) {
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

	var subjects = []string{}

	for _, rec := range recs {
		subject, ok := rec.Tags["mqttSubject"].(string)
		if !ok {
			log.Fatal("Error asserting type for mqttSubject")
		}
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
				i.currentStore.setCurrent(rec.ID, currentInput{Value: &currentItem})
			},
		)
		if !subscribeToken.WaitTimeout(timeout) {
			log.Printf("Unable to subscribe to %s", i.brokerAddr)
		}
		if subscribeToken.Error() != nil {
			log.Print(subscribeToken.Error())
		}
		log.Printf("Subscribed to %s", subject)

		subjects = append(subjects, subject)
	}

	defer func() {
		for _, subject := range subjects {
			mqttClient.Unsubscribe(subject)
			log.Printf("Unsubscribed from %s", subject)
		}
	}()

	// Blocks indefinitely
	block := make(chan int)
	<-block

	// TODO: Periodically write history into API

	// TODO: Should block indefinitely.
}
