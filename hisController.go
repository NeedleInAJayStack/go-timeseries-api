package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type hisController struct {
	db *gorm.DB
}

// GET /his/:pointId?start=...&end=...
// Note that start and end are in seconds since epoch (1970-01-01T00:00:00Z)
func (h hisController) getHis(w http.ResponseWriter, request *http.Request) {
	pointIdString := request.PathValue("pointId")
	pointId, err := uuid.Parse(pointIdString)
	if err != nil {
		log.Printf("Invalid UUID: %s", pointIdString)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err = request.ParseForm()
	if err != nil {
		log.Printf("Cannot parse form: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params := request.Form

	var sqlResult []his
	query := h.db.Where(&his{PointId: pointId})
	// TODO: Change start/end to ISO8601
	if params["start"] != nil {
		startStr := params["start"][0]
		start, err := strconv.ParseInt(startStr, 0, 64)
		if err != nil {
			log.Printf("Cannot parse time: %s", startStr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		query.Where("ts >= ?", time.Unix(start, 0))
	}
	if params["end"] != nil {
		endStr := params["end"][0]
		end, err := strconv.ParseInt(endStr, 0, 64)
		if err != nil {
			log.Printf("Cannot parse time: %s", endStr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		query.Where("ts < ?", time.Unix(end, 0))
	}
	err = query.Order("ts asc").Find(&sqlResult).Error
	if err != nil {
		log.Printf("SQL Error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httpResult := []apiHis{}
	for _, sqlRow := range sqlResult {
		httpResult = append(httpResult, apiHis(sqlRow))
	}

	httpJson, err := json.Marshal(httpResult)
	if err != nil {
		log.Printf("Cannot encode response JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(httpJson)
}

// POST /his/:pointId
// Note that start and end are in seconds since epoch (1970-01-01T00:00:00Z)
func (h hisController) postHis(writer http.ResponseWriter, request *http.Request) {
	pointIdString := request.PathValue("pointId")
	pointId, err := uuid.Parse(pointIdString)
	if err != nil {
		log.Printf("Invalid UUID: %s", pointIdString)
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	decoder := json.NewDecoder(request.Body)
	var hisItem apiHisItem
	err = decoder.Decode(&hisItem)
	if err != nil {
		log.Printf("Cannot decode request JSON: %s", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	his := his{
		PointId: pointId,
		Ts:      hisItem.Ts,
		Value:   hisItem.Value,
	}

	err = h.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pointId"}, {Name: "ts"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&his).Error
	if err != nil {
		log.Printf("SQL Error: %s", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

// DELETE /his/:pointId?start=...&end=...
// Note that start and end are in seconds since epoch (1970-01-01T00:00:00Z)
func (h hisController) deleteHis(writer http.ResponseWriter, request *http.Request) {
	pointIdString := request.PathValue("pointId")
	pointId, err := uuid.Parse(pointIdString)
	if err != nil {
		log.Printf("Invalid UUID: %s", pointIdString)
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	err = request.ParseForm()
	if err != nil {
		log.Printf("Cannot parse form: %s", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	params := request.Form

	var sqlResult []his
	query := h.db.Where(&his{PointId: pointId})
	// TODO: Change start/end to ISO8601
	if params["start"] != nil {
		startStr := params["start"][0]
		start, err := strconv.ParseInt(startStr, 0, 64)
		if err != nil {
			log.Printf("Cannot parse time: %s", startStr)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		query.Where("ts >= ?", time.Unix(start, 0))
	}
	if params["end"] != nil {
		endStr := params["end"][0]
		end, err := strconv.ParseInt(endStr, 0, 64)
		if err != nil {
			log.Printf("Cannot parse time: %s", endStr)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		query.Where("ts < ?", time.Unix(end, 0))
	}
	err = query.Delete(&sqlResult).Error
	if err != nil {
		log.Printf("SQL Error: %s", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
