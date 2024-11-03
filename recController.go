package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type recController struct {
	db *gorm.DB
}

// GET /recs
func (recController recController) getRecs(w http.ResponseWriter, r *http.Request) {
	var sqlResult []rec
	db := recController.db
	tag := r.URL.Query().Get("tag")
	if tag != "" {
		db = db.Where(datatypes.JSONQuery("tags").HasKey(tag))
	}
	err := db.Order("dis").Find(&sqlResult).Error
	if err != nil {
		log.Printf("SQL Error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httpResult := []apiRec{}
	for _, sqlRow := range sqlResult {
		httpResult = append(httpResult, apiRec(sqlRow))
	}

	httpJson, err := json.Marshal(httpResult)
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

	rec := rec(apiRec)

	err = recController.db.Create(&rec).Error
	if err != nil {
		log.Printf("SQL Error: %s", err)
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
	var recs []rec
	err := recController.db.Where(datatypes.JSONQuery("tags").HasKey("siteMeter")).Order("dis").Find(&recs).Error
	if err != nil {
		log.Printf("SQL Error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	apiRecs := []apiRec{}
	for _, rec := range recs {
		apiRecs = append(apiRecs, apiRec(rec))
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
	var rec rec
	err = recController.db.First(&rec, "id = ?", id).Error
	if err != nil {
		log.Printf("SQL Error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	apiRec := apiRec(rec)

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

	var rec rec
	err = recController.db.First(&rec, id).Error
	if err != nil {
		log.Printf("No rec with ID: %s", id)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if apiRec.Dis != nil {
		rec.Dis = apiRec.Dis
	}
	if apiRec.Unit != nil {
		rec.Unit = apiRec.Unit
	}

	err = recController.db.Save(&rec).Error
	if err != nil {
		log.Printf("Unable to update: %s", err)
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

	err = recController.db.Delete(&rec{}, id).Error
	if err != nil {
		log.Printf("Unable to delete: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
