package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type recController struct {
	store recStore
}

// GET /recs
func (recController recController) getRecs(w http.ResponseWriter, r *http.Request) {
	recs, err := recController.store.readRecs(r.URL.Query().Get("tag"))
	if err != nil {
		log.Printf("SQL Error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httpJson, err := json.Marshal(recs)
	if err != nil {
		log.Printf("Cannot encode response JSON: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(httpJson)
}

// POST /recs
func (recController recController) postRecs(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var apiRec apiRec
	err := decoder.Decode(&apiRec)
	if err != nil {
		log.Printf("Cannot decode request JSON: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = recController.store.createRec(apiRec)
	if err != nil {
		log.Printf("Storage Error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	recJSON, err := json.Marshal(apiRec)
	if err != nil {
		log.Printf("Cannot encode response JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(recJSON)
}

// GET /recs/tag/:tag
func (recController recController) getRecsByTag(w http.ResponseWriter, r *http.Request) {
	apiRecs, err := recController.store.readRecs("siteMeter")
	if err != nil {
		log.Printf("Storage Error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httpJson, err := json.Marshal(apiRecs)
	if err != nil {
		log.Printf("Cannot encode response JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(httpJson)
}

// GET /recs/:id
func (recController recController) getRec(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := uuid.Parse(idString)
	if err != nil {
		log.Printf("Invalid UUID: %s", idString)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	apiRec, err := recController.store.readRec(id)
	if err != nil {
		log.Printf("Storage Error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httpJson, err := json.Marshal(apiRec)
	if err != nil {
		log.Printf("Cannot encode response JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(httpJson)
}

// PUT /recs/:id
func (recController recController) putRec(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := uuid.Parse(idString)
	if err != nil {
		log.Printf("Invalid UUID: %s", idString)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var apiRec apiRec
	err = decoder.Decode(&apiRec)
	if err != nil {
		log.Printf("Cannot decode request JSON: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = recController.store.updateRec(id, apiRec)
	if err != nil {
		log.Printf("Storage Error: %s", id)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /recs/:id
func (recController recController) deleteRec(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := uuid.Parse(idString)
	if err != nil {
		log.Printf("Invalid UUID: %s", idString)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = recController.store.deleteRec(id)
	if err != nil {
		log.Printf("Unable to delete: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
