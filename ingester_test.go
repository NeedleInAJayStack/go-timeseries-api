package main

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type IngesterTestSuite struct {
	suite.Suite
	valueEmitter *mockValueEmitter
	ingester     *ingester
	currentStore currentStore
}

func TestIngesterTestSuite(t *testing.T) {
	suite.Run(t, new(IngesterTestSuite))
}

func (suite *IngesterTestSuite) SetupTest() {
	currentStore := newInMemoryCurrentStore()
	valueEmitter := mockValueEmitter{}
	ingester := newIngester(
		currentStore,
		&valueEmitter,
	)

	suite.valueEmitter = &valueEmitter
	suite.ingester = &ingester
	suite.currentStore = currentStore
}

func (suite *IngesterTestSuite) TestIngester() {
	// Setup some records
	rec1 := rec{
		ID: uuid.New(),
		Tags: map[string]interface{}{
			"mqttTopic": "test",
		},
	}
	rec2 := rec{
		ID: uuid.New(),
		Tags: map[string]interface{}{
			"mqttTopic": "test",
		},
	}
	rec3 := rec{
		ID: uuid.New(),
		Tags: map[string]interface{}{
			"mqttTopic": "test2",
		},
	}

	// Check that refresh adds all subscriptions and emit sets the current value
	suite.ingester.refreshSubscriptions(
		[]rec{
			rec1,
			rec2,
			rec3,
		},
	)
	suite.valueEmitter.emit(0.0)
	assert.Equal(suite.T(), suite.ingester.topics["test"][rec1.ID], true)
	assert.Equal(suite.T(), suite.ingester.topics["test"][rec2.ID], true)
	assert.Equal(suite.T(), suite.ingester.topics["test2"][rec3.ID], true)
	assert.Equal(suite.T(), *suite.currentStore.getCurrent(rec1.ID).Value, 0.0)
	assert.Equal(suite.T(), *suite.currentStore.getCurrent(rec2.ID).Value, 0.0)
	assert.Equal(suite.T(), *suite.currentStore.getCurrent(rec3.ID).Value, 0.0)

	// Check that removing some records removes subscriptions and emit no longer sets their current value
	suite.ingester.refreshSubscriptions(
		[]rec{
			rec1,
		},
	)
	suite.valueEmitter.emit(1.0)
	assert.Equal(suite.T(), suite.ingester.topics["test"][rec1.ID], true)
	assert.Equal(suite.T(), suite.ingester.topics["test"][rec2.ID], false)
	assert.Equal(suite.T(), suite.ingester.topics["test2"][rec3.ID], false)
	assert.Equal(suite.T(), *suite.currentStore.getCurrent(rec1.ID).Value, 1.0)
	assert.Equal(suite.T(), *suite.currentStore.getCurrent(rec2.ID).Value, 0.0)
	assert.Equal(suite.T(), *suite.currentStore.getCurrent(rec3.ID).Value, 0.0)
}
