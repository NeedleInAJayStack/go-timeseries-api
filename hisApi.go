package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func registerHis(db *gorm.DB) {

	// GET /his/:pointId?start=...&end=...
	// Note that start and end are in seconds since epoch (1970-01-01T00:00:00Z)
	http.HandleFunc("GET /his/{pointId}", func(w http.ResponseWriter, request *http.Request) {
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
		query := db.Where(&his{PointId: pointId})
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
		query.Order("ts asc").Find(&sqlResult)

		var httpResult []apiHis
		for _, sqlRow := range sqlResult {
			var value *float64
			if sqlRow.Value.Valid {
				value = &sqlRow.Value.Float64
			}
			httpResult = append(httpResult, apiHis{PointId: sqlRow.PointId, Ts: sqlRow.Ts, Value: value})
		}

		httpJson, err := json.Marshal(httpResult)
		if err != nil {
			log.Printf("Cannot encode response JSON")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(httpJson)
	})
}
