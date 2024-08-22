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
	// Test: curl -f 'http://localhost:8080/recs'
	// SQL seed: INSERT INTO rec VALUES ('424c159f-0eff-4a4d-8873-c2318c1809b1', '{"particleDeviceId": "abc", "particleVariableName": "eco2"}'::jsonb, '1211 Living Room CO2', 'ppm');
	var sqlResult []rec
	err := recController.db.Find(&sqlResult).Error
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
	// Test: curl -f --json '{ "id": "424c159f-0eff-4a4d-8873-c2318c1809b1", "tags": {"particleDeviceId": "abc", "particleVariableName": "eco2"}, "dis": "1211 Living Room CO2", "unit": "ppm" }' 'http://localhost:8080/recs'
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
	// Test: curl -f 'http://localhost:8080/recs/tag/particleDeviceId'
	tag := r.PathValue("tag")
	var recs []rec
	err := recController.db.Where(datatypes.JSONQuery("tags").HasKey(tag)).Order("dis").Find(&recs).Error
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
	// Test: curl -f 'http://localhost:8080/recs/424c159f-0eff-4a4d-8873-c2318c1809b1'
	idString := r.PathValue("id")
	id, err := uuid.Parse(idString)
	if err != nil {
		log.Printf("Invalid UUID: %s", idString)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var rec rec
	err = recController.db.First(&rec, id).Error
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
	// Test: curl -f -X PUT --json '{ "id": "424c159f-0eff-4a4d-8873-c2318c1809b1", "dis": "test" }' 'http://localhost:8080/recs/424c159f-0eff-4a4d-8873-c2318c1809b1'
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
	// Test: curl -f -X DELETE 'http://localhost:8080/recs/424c159f-0eff-4a4d-8873-c2318c1809b1'
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
