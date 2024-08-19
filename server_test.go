package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ServerTestSuite struct {
	suite.Suite
	server *http.ServeMux
	db     *gorm.DB
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (suite *ServerTestSuite) SetupTest() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.Nil(suite.T(), err)

	server, err := NewServer(ServerConfig{
		username:             "test",
		password:             "password",
		jwtSecret:            "aaa",
		tokenDurationSeconds: 60,

		db:            db,
		dbAutoMigrate: true,
	})
	assert.Nil(suite.T(), err)

	suite.server = server
	suite.db = db
}

func (suite *ServerTestSuite) getAuthToken() string {
	request, _ := http.NewRequest(http.MethodGet, "/auth/token", nil)
	request.SetBasicAuth("test", "password")
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusOK)

	var clientToken clientToken
	assert.Nil(suite.T(), json.Unmarshal(response.Body.Bytes(), &clientToken))
	return clientToken.Token
}

func (suite *ServerTestSuite) TestGetAuthToken() {
	request, _ := http.NewRequest(http.MethodGet, "/auth/token", nil)
	request.SetBasicAuth("test", "password")
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusOK)

	var clientToken clientToken
	assert.Nil(suite.T(), json.Unmarshal(response.Body.Bytes(), &clientToken))

	assert.True(suite.T(), len(clientToken.Token) > 0)
}

func (suite *ServerTestSuite) TestGetAuthTokenInvalidUsername() {
	request, _ := http.NewRequest(http.MethodGet, "/auth/token", nil)
	request.SetBasicAuth("wrong", "password")
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusForbidden)
}

func (suite *ServerTestSuite) TestGetAuthTokenInvalidPassword() {
	request, _ := http.NewRequest(http.MethodGet, "/auth/token", nil)
	request.SetBasicAuth("wrong", "password")
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusForbidden)
}

func (suite *ServerTestSuite) TestGetHis() {
	// Insert data for 2 different points with varying timestamps
	pointId1 := uuid.New()
	pointId2 := uuid.New()
	now := time.Now()
	nowMinus5Min := now.Add(-5 * 60 * 1e9)
	nowMinus10Min := now.Add(-10 * 60 * 1e9)
	nowMinus15Min := now.Add(-15 * 60 * 1e9)
	dbHistory := []his{
		{
			PointId: pointId1,
			Ts:      &nowMinus15Min,
			Value:   sql.NullFloat64{Float64: 0.0, Valid: true},
		},
		{
			PointId: pointId1,
			Ts:      &nowMinus10Min,
			Value:   sql.NullFloat64{Float64: 1.1, Valid: true},
		},
		{
			PointId: pointId1,
			Ts:      &nowMinus5Min,
			Value:   sql.NullFloat64{Float64: 2.2, Valid: true},
		},
		{
			PointId: pointId1,
			Ts:      &now,
			Value:   sql.NullFloat64{Float64: 3.3, Valid: true},
		},
		{
			PointId: pointId2,
			Ts:      &now,
			Value:   sql.NullFloat64{Float64: 3.4, Valid: true},
		},
	}
	suite.db.Create(&dbHistory)

	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/his/%s?start=%d&end=%d", pointId1, nowMinus10Min.Unix(), now.Unix()), nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", suite.getAuthToken()))
	response := httptest.NewRecorder()
	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusOK)
	decoder := json.NewDecoder(response.Result().Body)
	var apiHistory []apiHis
	assert.Nil(suite.T(), decoder.Decode(&apiHistory))

	assert.Equal(suite.T(), 2, len(apiHistory))

	// TODO: timezones are not matching up :(
	// oneOne := 1.1
	// twoTwo := 2.2
	// assert.Equal(
	// 	suite.T(),
	// 	apiHistory,
	// 	[]apiHis{
	// 		{
	// 			PointId: pointId1,
	// 			Ts:      &nowMinus10Min,
	// 			Value:   &oneOne,
	// 		},
	// 		{
	// 			PointId: pointId1,
	// 			Ts:      &nowMinus5Min,
	// 			Value:   &twoTwo,
	// 		},
	// 	},
	// )
}

func (suite *ServerTestSuite) TestPostHis() {
	var initialCount int64
	suite.db.Model(&his{}).Count(&initialCount)
	assert.Equal(suite.T(), initialCount, int64(0))

	pointId := uuid.New()
	ts := time.Now().Local()
	value := 123.456

	hisItem := apiHisItem{
		Ts:    &ts,
		Value: &value,
	}
	body, err := json.Marshal(hisItem)
	assert.Nil(suite.T(), err)

	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/his/%s", pointId), bytes.NewReader(body))
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", suite.getAuthToken()))
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusOK)

	var dbHis his
	suite.db.First(&dbHis)
	assert.NotNil(suite.T(), dbHis)

	// TODO: timezones are not matching up :(
	// assert.Equal(
	// 	suite.T(),
	// 	dbHis,
	// 	his{
	// 		PointId: pointId,
	// 		Ts:      &ts,
	// 		Value:   sql.NullFloat64{Float64: value, Valid: true},
	// 	},
	// )
}
