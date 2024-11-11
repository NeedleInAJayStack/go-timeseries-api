package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type hisController struct {
	store historyStore
}

// GET /recs/:pointId/history?start=...&end=...
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

	// TODO: Change start/end to ISO8601
	var start *time.Time
	if params["start"] != nil {
		startStr := params["start"][0]
		startUnix, err := strconv.ParseInt(startStr, 0, 64)
		if err != nil {
			log.Printf("Cannot parse time: %s", startStr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		startTime := time.Unix(startUnix, 0)
		start = &startTime
	}
	var end *time.Time
	if params["end"] != nil {
		endStr := params["end"][0]
		endUnix, err := strconv.ParseInt(endStr, 0, 64)
		if err != nil {
			log.Printf("Cannot parse time: %s", endStr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		endTime := time.Unix(endUnix, 0)
		end = &endTime
	}
	httpResult, err := h.store.readHistory(pointId, start, end)
	if err != nil {
		log.Printf("Storage Error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
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

// POST /recs/:pointId/history
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
	var hisItem hisItem
	err = decoder.Decode(&hisItem)
	if err != nil {
		log.Printf("Cannot decode request JSON: %s", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.store.writeHistory(pointId, hisItem)
	if err != nil {
		log.Printf("Storage Error: %s", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

// DELETE /recs/:pointId/history?start=...&end=...
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

	// TODO: Change start/end to ISO8601
	var start *time.Time
	if params["start"] != nil {
		startStr := params["start"][0]
		startUnix, err := strconv.ParseInt(startStr, 0, 64)
		if err != nil {
			log.Printf("Cannot parse time: %s", startStr)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		startTime := time.Unix(startUnix, 0)
		start = &startTime
	}
	var end *time.Time
	if params["end"] != nil {
		endStr := params["end"][0]
		endUnix, err := strconv.ParseInt(endStr, 0, 64)
		if err != nil {
			log.Printf("Cannot parse time: %s", endStr)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		endTime := time.Unix(endUnix, 0)
		end = &endTime
	}
	err = h.store.deleteHistory(pointId, start, end)
	if err != nil {
		log.Printf("SQL Error: %s", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
