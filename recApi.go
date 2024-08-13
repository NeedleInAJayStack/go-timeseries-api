package main

import (
	"encoding/json"
	"log"
	"net/http"

	"gorm.io/gorm"
)

func registerRecs(db *gorm.DB) {

	// GET /recs
	http.HandleFunc("GET /recs", func(w http.ResponseWriter, r *http.Request) {
		// Test: curl -f 'http://localhost:8080/recs'
		// SQL seed: INSERT INTO rec VALUES ('424c159f-0eff-4a4d-8873-c2318c1809b1', '{"particleDeviceId": "abc", "particleVariableName": "eco2"}'::jsonb, '1211 Living Room CO2', 'ppm');
		var sqlResult []rec
		err := db.Find(&sqlResult).Error
		if err != nil {
			log.Printf("SQL Error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		httpResult := []apiRec{}
		for _, sqlRow := range sqlResult {
			var dis *string
			if sqlRow.Dis.Valid {
				dis = &sqlRow.Dis.String
			}
			var unit *string
			if sqlRow.Unit.Valid {
				unit = &sqlRow.Unit.String
			}
			httpResult = append(httpResult, apiRec{ID: sqlRow.ID, Tags: sqlRow.Tags, Dis: dis, Unit: unit})
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
