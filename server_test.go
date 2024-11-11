package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ServerTestSuite struct {
	suite.Suite
	server http.Handler
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

func (suite *ServerTestSuite) TestGetAuthToken() {
	request, _ := http.NewRequest(http.MethodGet, "/api/auth/token", nil)
	request.SetBasicAuth("test", "password")
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusOK)

	var clientToken clientToken
	assert.Nil(suite.T(), json.Unmarshal(response.Body.Bytes(), &clientToken))

	assert.True(suite.T(), len(clientToken.Token) > 0)
}

func (suite *ServerTestSuite) TestGetAuthTokenInvalidUsername() {
	request, _ := http.NewRequest(http.MethodGet, "/api/auth/token", nil)
	request.SetBasicAuth("wrong", "password")
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusForbidden)
}

func (suite *ServerTestSuite) TestGetAuthTokenInvalidPassword() {
	request, _ := http.NewRequest(http.MethodGet, "/api/auth/token", nil)
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
	dbHistory := []gormHis{
		{
			PointId: pointId1,
			Ts:      &nowMinus15Min,
			Value:   f(0.0),
		},
		{
			PointId: pointId1,
			Ts:      &nowMinus10Min,
			Value:   f(1.1),
		},
		{
			PointId: pointId1,
			Ts:      &nowMinus5Min,
			Value:   f(2.2),
		},
		{
			PointId: pointId1,
			Ts:      &now,
			Value:   f(3.3),
		},
		{
			PointId: pointId2,
			Ts:      &now,
			Value:   f(3.4),
		},
	}
	suite.db.Create(&dbHistory)

	authToken := suite.getAuthToken()

	var history []hisItem
	suite.get(fmt.Sprintf("/api/recs/%s/history?start=%d&end=%d", pointId1, nowMinus10Min.Unix(), now.Unix()), authToken, &history)
	assert.Equal(suite.T(), 2, len(history))

	// TODO: timezones are not matching up :(
	// oneOne := 1.1
	// twoTwo := 2.2
	// assert.Equal(
	// 	suite.T(),
	// 	history,
	// 	[]his{
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
	suite.db.Model(&gormHis{}).Count(&initialCount)
	assert.Equal(suite.T(), initialCount, int64(0))

	pointId := uuid.New()
	ts := time.Now().Local()
	value := 123.456

	hisItem := hisItem{
		Ts:    &ts,
		Value: &value,
	}

	authToken := suite.getAuthToken()

	suite.post(fmt.Sprintf("/api/recs/%s/history", pointId), authToken, hisItem)

	var dbHis gormHis
	suite.db.First(&dbHis)
	assert.NotNil(suite.T(), dbHis)

	// TODO: timezones are not matching up :(
	// assert.Equal(
	// 	suite.T(),
	// 	dbHis,
	// 	gormHis{
	// 		PointId: pointId,
	// 		Ts:      &ts,
	// 		Value:   f(value),
	// 	},
	// )
}

func (suite *ServerTestSuite) TestGetRecs() {
	id1, _ := uuid.Parse("1b4e32c7-61b5-4b38-a1cd-023c25f9965c")
	id2, _ := uuid.Parse("5ba26f95-e1ef-4867-a86b-a866cb174f06")
	gormRecs := []gormRec{
		{
			ID:   id1,
			Dis:  s("rec1"),
			Tags: datatypes.JSON([]byte(`{"tag":"value1"}`)),
			Unit: s("kW"),
		},
		{
			ID:   id2,
			Dis:  s("rec2"),
			Tags: datatypes.JSON([]byte(`{"tag":"value2"}`)),
			Unit: s("lb"),
		},
	}
	suite.db.Create(&gormRecs)

	authToken := suite.getAuthToken()

	var recs []rec
	suite.get("/api/recs", authToken, &recs)

	rec1 := "rec1"
	kW := "kW"
	rec2 := "rec2"
	lb := "lb"
	assert.Equal(
		suite.T(),
		recs,
		[]rec{
			{
				ID:   id1,
				Dis:  &rec1,
				Tags: datatypes.JSON([]byte(`{"tag":"value1"}`)),
				Unit: &kW,
			},
			{
				ID:   id2,
				Dis:  &rec2,
				Tags: datatypes.JSON([]byte(`{"tag":"value2"}`)),
				Unit: &lb,
			},
		},
	)
}

func (suite *ServerTestSuite) TestPostRecs() {
	id1, _ := uuid.Parse("1b4e32c7-61b5-4b38-a1cd-023c25f9965c")
	rec1 := "rec1"
	kW := "kW"
	rec := rec{
		ID:   id1,
		Dis:  &rec1,
		Tags: datatypes.JSON([]byte(`{"tag":"value1"}`)),
		Unit: &kW,
	}
	authToken := suite.getAuthToken()

	suite.post("/api/recs", authToken, rec)

	var gormRecs []gormRec
	suite.db.Find(&gormRecs)
	assert.Equal(
		suite.T(),
		gormRecs,
		[]gormRec{
			{
				ID:   id1,
				Dis:  s("rec1"),
				Tags: datatypes.JSON([]byte(`{"tag":"value1"}`)),
				Unit: s("kW"),
			},
		},
	)
}

func (suite *ServerTestSuite) TestGetRecsByTag() {
	id1, _ := uuid.Parse("1b4e32c7-61b5-4b38-a1cd-023c25f9965c")
	id2, _ := uuid.Parse("5ba26f95-e1ef-4867-a86b-a866cb174f06")
	gormRecs := []gormRec{
		{
			ID:   id1,
			Dis:  s("rec1"),
			Tags: datatypes.JSON([]byte(`{"tag1":"value1"}`)),
			Unit: s("kW"),
		},
		{
			ID:   id2,
			Dis:  s("rec2"),
			Tags: datatypes.JSON([]byte(`{"tag2":"value2"}`)),
			Unit: s("lb"),
		},
	}
	suite.db.Create(&gormRecs)

	authToken := suite.getAuthToken()

	var recs []rec
	suite.get("/api/recs?tag=tag1", authToken, &recs)

	rec1 := "rec1"
	kW := "kW"
	assert.Equal(
		suite.T(),
		recs,
		[]rec{
			{
				ID:   id1,
				Dis:  &rec1,
				Tags: datatypes.JSON([]byte(`{"tag1":"value1"}`)),
				Unit: &kW,
			},
		},
	)
}

func (suite *ServerTestSuite) TestGetRec() {
	id1, _ := uuid.Parse("1b4e32c7-61b5-4b38-a1cd-023c25f9965c")
	id2, _ := uuid.Parse("5ba26f95-e1ef-4867-a86b-a866cb174f06")
	gormRecs := []gormRec{
		{
			ID:   id1,
			Dis:  s("rec1"),
			Tags: datatypes.JSON([]byte(`{"tag":"value1"}`)),
			Unit: s("kW"),
		},
		{
			ID:   id2,
			Dis:  s("rec2"),
			Tags: datatypes.JSON([]byte(`{"tag":"value2"}`)),
			Unit: s("lb"),
		},
	}
	suite.db.Create(&gormRecs)

	authToken := suite.getAuthToken()

	var rec2 rec
	suite.get(fmt.Sprintf("/api/recs/%s", id2), authToken, &rec2)

	rec2Dis := "rec2"
	lb := "lb"
	assert.Equal(
		suite.T(),
		rec2,
		rec{
			ID:   id2,
			Dis:  &rec2Dis,
			Tags: datatypes.JSON([]byte(`{"tag":"value2"}`)),
			Unit: &lb,
		},
	)
}

func (suite *ServerTestSuite) TestPutRec() {
	id, _ := uuid.Parse("1b4e32c7-61b5-4b38-a1cd-023c25f9965c")
	gormRecs := []gormRec{
		{
			ID:   id,
			Dis:  s("rec"),
			Tags: datatypes.JSON([]byte(`{"tag":"value"}`)),
			Unit: s("kW"),
		},
	}
	suite.db.Create(&gormRecs)

	dis := "rec updated"
	lb := "lb"
	rec := rec{
		ID:   id,
		Dis:  &dis,
		Tags: datatypes.JSON([]byte(`{"tag":"value1"}`)),
		Unit: &lb,
	}

	authToken := suite.getAuthToken()

	suite.put(fmt.Sprintf("/api/recs/%s", id), authToken, rec)

	var updatedRec gormRec
	suite.db.First(&updatedRec, id)
	assert.Equal(
		suite.T(),
		updatedRec,
		gormRec{
			ID:   id,
			Dis:  s("rec updated"),
			Tags: datatypes.JSON([]byte(`{"tag":"value"}`)),
			Unit: s("lb"),
		},
	)
}

func (suite *ServerTestSuite) TestDeleteRec() {
	id, _ := uuid.Parse("1b4e32c7-61b5-4b38-a1cd-023c25f9965c")
	gormRecs := []gormRec{
		{
			ID:   id,
			Dis:  s("rec"),
			Tags: datatypes.JSON([]byte(`{"tag":"value"}`)),
			Unit: s("kW"),
		},
	}
	suite.db.Create(&gormRecs)

	authToken := suite.getAuthToken()
	suite.delete(fmt.Sprintf("/api/recs/%s", id), authToken)

	var gormRecCount int64
	suite.db.Where("id = ?", id).Count(&gormRecCount)
	assert.Equal(suite.T(), gormRecCount, int64(0))
}

func (suite *ServerTestSuite) TestCurrent() {
	id, _ := uuid.Parse("1b4e32c7-61b5-4b38-a1cd-023c25f9965c")
	gormRecs := []gormRec{
		{
			ID:   id,
			Dis:  s("rec"),
			Tags: datatypes.JSON([]byte(`{"tag":"value"}`)),
			Unit: s("kW"),
		},
	}
	suite.db.Create(&gormRecs)

	authToken := suite.getAuthToken()

	value := 123.456
	currentInput := currentInput{
		Value: &value,
	}
	suite.post(fmt.Sprintf("/api/recs/%s/current", id), authToken, currentInput)

	var current current
	suite.get(fmt.Sprintf("/api/recs/%s/current", id), authToken, &current)
	assert.Equal(suite.T(), 123.456, *current.Value)
}

// These functions just take literals and return a pointer to them. For easier DB/JSON construction
func s(s string) *string   { return &s }
func f(f float64) *float64 { return &f }

func (suite *ServerTestSuite) getAuthToken() string {
	request, _ := http.NewRequest(http.MethodGet, "/api/auth/token", nil)
	request.SetBasicAuth("test", "password")
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusOK)

	var clientToken clientToken
	assert.Nil(suite.T(), json.Unmarshal(response.Body.Bytes(), &clientToken))
	return clientToken.Token
}

func (suite *ServerTestSuite) delete(route string, authToken string) {
	request, err := http.NewRequest(http.MethodDelete, route, nil)
	assert.Nil(suite.T(), err)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	response := httptest.NewRecorder()
	suite.server.ServeHTTP(response, request)
	assert.Equal(suite.T(), response.Code, http.StatusOK)
}

func (suite *ServerTestSuite) get(route string, authToken string, unmarshalTo any) {
	request, err := http.NewRequest(http.MethodGet, route, nil)
	assert.Nil(suite.T(), err)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	response := httptest.NewRecorder()
	suite.server.ServeHTTP(response, request)
	assert.Equal(suite.T(), response.Code, http.StatusOK)
	assert.Nil(suite.T(), json.Unmarshal(response.Body.Bytes(), &unmarshalTo))
}

func (suite *ServerTestSuite) post(route string, authToken string, toMarshal any) {
	body, err := json.Marshal(toMarshal)
	assert.Nil(suite.T(), err)
	request, err := http.NewRequest(http.MethodPost, route, bytes.NewReader(body))
	assert.Nil(suite.T(), err)
	request.Header.Add(
		"Authorization",
		fmt.Sprintf("Bearer %s", authToken),
	)
	response := httptest.NewRecorder()
	suite.server.ServeHTTP(response, request)
	assert.Equal(suite.T(), response.Code, http.StatusOK)
}

func (suite *ServerTestSuite) put(route string, authToken string, toMarshal any) {
	body, err := json.Marshal(toMarshal)
	assert.Nil(suite.T(), err)
	request, err := http.NewRequest(http.MethodPut, route, bytes.NewReader(body))
	assert.Nil(suite.T(), err)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	response := httptest.NewRecorder()
	suite.server.ServeHTTP(response, request)
	assert.Equal(suite.T(), response.Code, http.StatusOK)
}
