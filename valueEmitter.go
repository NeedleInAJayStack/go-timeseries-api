package main

import (
	"encoding/json"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type valueEmitter interface {
	subscribe(source string, onEvent func(string, float64))
	unsubscribe(source string)
}

type mqttValueEmitter struct {
	connectionTimeout time.Duration
	subscribeTimeout  time.Duration

	mqttClient mqtt.Client
}

func newMQTTValueEmitter(
	mqttClient mqtt.Client,
) mqttValueEmitter {
	return mqttValueEmitter{
		// TODO: Make configurable
		connectionTimeout: time.Duration(5 * time.Second),
		subscribeTimeout:  time.Duration(5 * time.Second),
		mqttClient:        mqttClient,
	}
}

func (m *mqttValueEmitter) subscribe(source string, onEvent func(string, float64)) {
	subscribeToken := m.mqttClient.Subscribe(source, 0, func(c mqtt.Client, m mqtt.Message) {
		var value float64
		err := json.Unmarshal(m.Payload(), &value)
		if err != nil {
			log.Printf("Cannot decode message JSON: %s", err)
			return
		}
		onEvent(source, value)
	})
	if !subscribeToken.WaitTimeout(m.subscribeTimeout) {
		log.Printf("Unable to subscribe to %s", source)
	}
	if subscribeToken.Error() != nil {
		log.Print(subscribeToken.Error())
	}
	log.Printf("Subscribed to %s", source)
}

func (m *mqttValueEmitter) unsubscribe(source string) {
	unsubscribeToken := m.mqttClient.Unsubscribe(source)
	if !unsubscribeToken.WaitTimeout(m.subscribeTimeout) {
		log.Printf("Unable to unsubscribe to %s", source)
	}
	if unsubscribeToken.Error() != nil {
		log.Print(unsubscribeToken.Error())
	}
	log.Printf("Unsubscribed from %s", source)
}

// For testing
type mockValueEmitter struct {
	sources []string
	onEvent func(string, float64)
}

func (m *mockValueEmitter) subscribe(source string, onEvent func(string, float64)) {
	m.sources = append(m.sources, source)
	// Just store the last one, since they should all be the same.
	m.onEvent = onEvent
}

func (m *mockValueEmitter) unsubscribe(source string) {
	// Do nothing
}

func (m *mockValueEmitter) emit(value float64) {
	for _, source := range m.sources {
		m.onEvent(source, value)
	}
}
