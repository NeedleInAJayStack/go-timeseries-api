package main

import (
	"log"
	"sync"

	"github.com/google/uuid"
)

type ingester struct {
	currentStore currentStore
	valueEmitter valueEmitter

	// Mutable
	// Stores a list of topic names that have been subscribed to.
	topics map[string]map[uuid.UUID]bool
	mux    *sync.RWMutex
}

func newIngester(
	currentStore currentStore,
	valueEmitter valueEmitter,
) ingester {
	return ingester{
		currentStore: currentStore,
		valueEmitter: valueEmitter,

		topics: map[string]map[uuid.UUID]bool{},
		mux:    &sync.RWMutex{},
	}
}

func (i *ingester) refreshSubscriptions(recs []rec) {
	// Use read-write lock to avoid modifying topics while onMessage is processing.
	// TODO: Consider breaking out a separate data structure type for topics.
	i.mux.Lock()
	defer i.mux.Unlock()

	recIDs := map[uuid.UUID]bool{}
	for _, record := range recs {
		recIDs[record.ID] = true
	}

	toSubscribe := []topicAndId{}
	for _, record := range recs {
		topic, ok := record.Tags["mqttTopic"].(string)
		if !ok {
			log.Printf("Error asserting type for mqttTopic")
			continue
		}

		_, present := i.topics[topic]
		if !present {
			toSubscribe = append(toSubscribe, topicAndId{topic: topic, recID: record.ID})
			continue
		}
		_, present = i.topics[topic][record.ID]
		if !present {
			toSubscribe = append(toSubscribe, topicAndId{topic: topic, recID: record.ID})
			continue
		}
	}

	toUnsubscribe := []topicAndId{}
	for topic, subscribedRecIDs := range i.topics {
		for subscribedRecID, _ := range subscribedRecIDs {
			_, present := recIDs[subscribedRecID]
			if !present {
				toUnsubscribe = append(toUnsubscribe, topicAndId{topic: topic, recID: subscribedRecID})
			}
		}
	}

	for _, topicAndId := range toSubscribe {
		i.subscribe(topicAndId.topic, topicAndId.recID)
	}
	for _, topicAndId := range toUnsubscribe {
		i.unsubscribe(topicAndId.topic, topicAndId.recID)
	}
}

// Helper methods

// Subscribe to a topic, and associate the rec with the topic
func (i *ingester) subscribe(topic string, recID uuid.UUID) {
	i.valueEmitter.subscribe(
		topic,
		func(source string, value float64) {
			// Ensure that we are not modifying topics while onMessage is processing.
			i.mux.RLock()
			defer i.mux.RUnlock()

			recIDs := i.topics[source]
			for recID, _ := range recIDs {
				i.currentStore.setCurrent(recID, currentInput{Value: &value})
			}
		},
	)

	_, present := i.topics[topic]
	if present {
		i.topics[topic][recID] = true
	} else {
		i.topics[topic] = map[uuid.UUID]bool{recID: true}
	}
}

func (i *ingester) unsubscribe(topic string, recID uuid.UUID) {
	if i.topics[topic] == nil {
		return
	} else {
		delete(i.topics[topic], recID)
	}

	if len(i.topics[topic]) == 0 {
		i.valueEmitter.unsubscribe(topic)
	}
}

type topicAndId struct {
	topic string
	recID uuid.UUID
}
