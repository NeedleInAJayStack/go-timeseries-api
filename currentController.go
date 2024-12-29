package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type currentController struct {
	store currentStore
}

// GET /recs/:pointId/current
// Note that start and end are in seconds since epoch (1970-01-01T00:00:00Z)
func (h currentController) getCurrent(w http.ResponseWriter, request *http.Request) {
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

	httpResult, err := h.store.getCurrent(pointId)
	if err != nil {
		log.Printf("Cannot retrieve current value")
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

// POST /recs/:pointId/current
// Note that start and end are in seconds since epoch (1970-01-01T00:00:00Z)
func (h currentController) postCurrent(writer http.ResponseWriter, request *http.Request) {
	pointIdString := request.PathValue("pointId")
	pointId, err := uuid.Parse(pointIdString)
	if err != nil {
		log.Printf("Invalid UUID: %s", pointIdString)
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	decoder := json.NewDecoder(request.Body)
	var currentItem currentInput
	err = decoder.Decode(&currentItem)
	if err != nil {
		log.Printf("Cannot decode request JSON: %s", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.store.setCurrent(pointId, currentItem)
	if err != nil {
		log.Printf("Cannot save current value")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
