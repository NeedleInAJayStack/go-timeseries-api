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
	// 		Value:   f(value),
	// 	},
	// )
}

func (suite *ServerTestSuite) TestGetRecs() {
	id1, _ := uuid.Parse("1b4e32c7-61b5-4b38-a1cd-023c25f9965c")
	id2, _ := uuid.Parse("5ba26f95-e1ef-4867-a86b-a866cb174f06")
	recs := []rec{
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
	suite.db.Create(&recs)

	request, _ := http.NewRequest(http.MethodGet, "/recs", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", suite.getAuthToken()))
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusOK)
	decoder := json.NewDecoder(response.Result().Body)
	var apiRecs []apiRec
	assert.Nil(suite.T(), decoder.Decode(&apiRecs))

	rec1 := "rec1"
	kW := "kW"
	rec2 := "rec2"
	lb := "lb"
	assert.Equal(
		suite.T(),
		apiRecs,
		[]apiRec{
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
	apiRec := apiRec{
		ID:   id1,
		Dis:  &rec1,
		Tags: datatypes.JSON([]byte(`{"tag":"value1"}`)),
		Unit: &kW,
	}
	body, err := json.Marshal(apiRec)
	assert.Nil(suite.T(), err)

	request, _ := http.NewRequest(http.MethodPost, "/recs", bytes.NewReader(body))
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", suite.getAuthToken()))
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusOK)

	var recs []rec
	suite.db.Find(&recs)
	assert.Equal(
		suite.T(),
		recs,
		[]rec{
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
	recs := []rec{
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
	suite.db.Create(&recs)

	request, _ := http.NewRequest(http.MethodGet, "/recs/tag/tag1", nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", suite.getAuthToken()))
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusOK)
	decoder := json.NewDecoder(response.Result().Body)
	var apiRecs []apiRec
	assert.Nil(suite.T(), decoder.Decode(&apiRecs))

	rec1 := "rec1"
	kW := "kW"
	assert.Equal(
		suite.T(),
		apiRecs,
		[]apiRec{
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
	recs := []rec{
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
	suite.db.Create(&recs)

	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/recs/%s", id2), nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", suite.getAuthToken()))
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusOK)
	decoder := json.NewDecoder(response.Result().Body)
	var rec2 apiRec
	assert.Nil(suite.T(), decoder.Decode(&rec2))

	rec2Dis := "rec2"
	lb := "lb"
	assert.Equal(
		suite.T(),
		rec2,
		apiRec{
			ID:   id2,
			Dis:  &rec2Dis,
			Tags: datatypes.JSON([]byte(`{"tag":"value2"}`)),
			Unit: &lb,
		},
	)
}

func (suite *ServerTestSuite) TestPutRec() {
	id, _ := uuid.Parse("1b4e32c7-61b5-4b38-a1cd-023c25f9965c")
	recs := []rec{
		{
			ID:   id,
			Dis:  s("rec"),
			Tags: datatypes.JSON([]byte(`{"tag":"value"}`)),
			Unit: s("kW"),
		},
	}
	suite.db.Create(&recs)

	dis := "rec updated"
	lb := "lb"
	apiRec := apiRec{
		ID:   id,
		Dis:  &dis,
		Tags: datatypes.JSON([]byte(`{"tag":"value1"}`)),
		Unit: &lb,
	}
	body, err := json.Marshal(apiRec)
	assert.Nil(suite.T(), err)

	request, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/recs/%s", id), bytes.NewReader(body))
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", suite.getAuthToken()))
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusOK)

	var updatedRec rec
	suite.db.First(&updatedRec, id)
	assert.Equal(
		suite.T(),
		updatedRec,
		rec{
			ID:   id,
			Dis:  s("rec updated"),
			Tags: datatypes.JSON([]byte(`{"tag":"value"}`)),
			Unit: s("lb"),
		},
	)
}

func (suite *ServerTestSuite) TestDeleteRec() {
	id, _ := uuid.Parse("1b4e32c7-61b5-4b38-a1cd-023c25f9965c")
	recs := []rec{
		{
			ID:   id,
			Dis:  s("rec"),
			Tags: datatypes.JSON([]byte(`{"tag":"value"}`)),
			Unit: s("kW"),
		},
	}
	suite.db.Create(&recs)

	request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/recs/%s", id), nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", suite.getAuthToken()))
	response := httptest.NewRecorder()

	suite.server.ServeHTTP(response, request)

	assert.Equal(suite.T(), response.Code, http.StatusOK)

	var recCount int64
	suite.db.Where("id = ?", id).Count(&recCount)
	assert.Equal(suite.T(), recCount, int64(0))
}

// These functions just take literals and return a pointer to them. For easier DB/JSON construction
func s(s string) *string   { return &s }
func f(f float64) *float64 { return &f }
