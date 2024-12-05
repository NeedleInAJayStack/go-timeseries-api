package main

import (
	"encoding/json"
	"fmt"
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

	// Mutable
	// Stores a list of subject names that have been subscribed to.
	subjects   map[string][]rec
	mqttClient mqtt.Client
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

func (i mqttIngester) start() error {
	err := i.connect()
	if err != nil {
		return err
	}

	// TODO: Run periodically so that points may be added/removed
	err = i.fetchSubjects()
	if err != nil {
		return err
	}

	i.subscribe()
	return nil
}

func (i mqttIngester) stop() {
	for subject, _ := range i.subjects {
		i.mqttClient.Unsubscribe(subject)
		i.subjects[subject] = nil
		log.Printf("Unsubscribed from %s", subject)
	}
	i.disconnect()
}

// Helper methods

func (i mqttIngester) connect() error {
	i.mqttClient = mqtt.NewClient(mqtt.NewClientOptions().AddBroker(i.brokerAddr))
	connectToken := i.mqttClient.Connect()
	if !connectToken.WaitTimeout(i.connectionTimeout) {
		return fmt.Errorf("unable to connect to %s", i.brokerAddr)
	}
	if connectToken.Error() != nil {
		return connectToken.Error()
	}
	log.Printf("Connected to %s", i.brokerAddr)
	return nil
}

func (i mqttIngester) disconnect() {
	i.mqttClient.Disconnect(1)
	log.Printf("Disconnected from %s", i.brokerAddr)
}

func (i mqttIngester) fetchSubjects() error {
	recs, err := i.recStore.readRecs("mqttSubject")
	if err != nil {
		return fmt.Errorf("error getting mqttSubject points: %s", err)
	}
	for _, record := range recs {
		subject, ok := record.Tags["mqttSubject"].(string)
		if !ok {
			log.Printf("Error asserting type for mqttSubject")
		}
		if i.subjects[subject] != nil {
			i.subjects[subject] = append(i.subjects[subject], record)
		} else {
			i.subjects[subject] = []rec{record}
		}
	}
	return nil
}

// Subscribe to each subject
func (i mqttIngester) subscribe() {
	for subject, recs := range i.subjects {
		subscribeToken := i.mqttClient.Subscribe(
			subject,
			0,
			func(c mqtt.Client, m mqtt.Message) {
				// TODO: Change messages to JSON
				// var currentItem currentInput
				var currentItem float64
				err := json.Unmarshal(m.Payload(), &currentItem)
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
}
