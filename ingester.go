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
	mqttClient   mqtt.Client
}

func (i mqttIngester) start() {
	recs, err := i.recStore.readRecs("mqttSubject")
	if err != nil {
		log.Fatalf("Error getting mqttSubject points: %s", err)
		return
	}

	// Configure
	// TODO: Add env var integration
	brokerAddr := "tcp://JaysDesktop.local:1883"
	timeout := time.Duration(5 * time.Second)

	connectToken := i.mqttClient.Connect()
	if !connectToken.WaitTimeout(timeout) {
		log.Fatalf("Unable to connect to %s", brokerAddr)
	}
	if connectToken.Error() != nil {
		log.Fatal(connectToken.Error())
	}
	log.Printf("Connected to %s", brokerAddr)
	defer func() {
		i.mqttClient.Disconnect(1)
		log.Printf("Disconnected from %s", brokerAddr)
	}()

	var subjects = []string{}

	for _, rec := range recs {
		subject, ok := rec.Tags["mqttSubject"].(string)
		if !ok {
			log.Fatal("Error asserting type for mqttSubject")
		}
		subscribeToken := i.mqttClient.Subscribe(
			subject,
			0,
			func(c mqtt.Client, m mqtt.Message) {
				var currentItem apiCurrentInput
				err = json.Unmarshal(m.Payload(), &currentItem)
				if err != nil {
					log.Printf("Cannot decode message JSON: %s", err)
					return
				}
				i.currentStore.setCurrent(rec.ID, currentItem)
			},
		)
		if !subscribeToken.WaitTimeout(timeout) {
			log.Printf("Unable to subscribe to %s", brokerAddr)
		}
		if subscribeToken.Error() != nil {
			log.Print(subscribeToken.Error())
		}
		log.Printf("Subscribed to %s", subject)

		subjects = append(subjects, subject)
	}

	defer func() {
		for _, subject := range subjects {
			i.mqttClient.Unsubscribe(subject)
			log.Printf("Unsubscribed from %s", subject)
		}
	}()

	// TODO: Periodically write history into API

	// TODO: Should block indefinitely.
}
