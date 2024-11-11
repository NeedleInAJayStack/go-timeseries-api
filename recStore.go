package main

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type recStore interface {
	readRecs(string) ([]apiRec, error)
	readRec(uuid.UUID) (*apiRec, error)
	createRec(apiRec) error
	updateRec(uuid.UUID, apiRec) error
	deleteRec(uuid.UUID) error
}

type apiRec struct {
	ID   uuid.UUID      `json:"id"`
	Tags datatypes.JSON `json:"tags"`
	Dis  *string        `json:"dis"`
	Unit *string        `json:"unit"`
}

// gormHistoryStore stores point historical values in a GORM database.
type gormRecStore struct {
	db *gorm.DB
}

func newGormRecStore(db *gorm.DB) gormRecStore {
	return gormRecStore{db: db}
}

func (s gormRecStore) readRecs(
	tag string,
) ([]apiRec, error) {
	var sqlResult []rec
	db := s.db
	if tag != "" {
		db = db.Where(datatypes.JSONQuery("tags").HasKey(tag))
	}
	err := db.Order("dis").Find(&sqlResult).Error
	if err != nil {
		return []apiRec{}, err
	}

	result := []apiRec{}
	for _, sqlRow := range sqlResult {
		result = append(result, apiRec(sqlRow))
	}
	return result, nil
}

func (s gormRecStore) readRec(
	id uuid.UUID,
) (*apiRec, error) {
	var rec rec
	err := s.db.First(&rec, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	apiRec := apiRec(rec)
	return &apiRec, nil
}

func (s gormRecStore) createRec(
	apiRec apiRec,
) error {
	rec := rec(apiRec)
	return s.db.Create(&rec).Error
}

func (s gormRecStore) updateRec(
	id uuid.UUID,
	apiRec apiRec,
) error {
	var rec rec
	err := s.db.First(&rec, id).Error
	if err != nil {
		return err
	}

	if apiRec.Dis != nil {
		rec.Dis = apiRec.Dis
	}
	if apiRec.Unit != nil {
		rec.Unit = apiRec.Unit
	}

	return s.db.Save(&rec).Error
}

func (s gormRecStore) deleteRec(id uuid.UUID) error {
	return s.db.Delete(&rec{}, id).Error
}

type rec struct {
	ID   uuid.UUID      `gorm:"column:id;type:uuid;primaryKey:rec_pkey"`
	Tags datatypes.JSON `gorm:"type:json"`
	Dis  *string
	Unit *string
}

func (rec) TableName() string {
	return "rec"
}
