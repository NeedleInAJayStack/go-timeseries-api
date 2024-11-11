package main

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type recStore interface {
	readRecs(string) ([]rec, error)
	readRec(uuid.UUID) (*rec, error)
	createRec(rec) error
	updateRec(uuid.UUID, rec) error
	deleteRec(uuid.UUID) error
}

type rec struct {
	ID   uuid.UUID         `json:"id"`
	Tags datatypes.JSONMap `json:"tags"`
	Dis  *string           `json:"dis"`
	Unit *string           `json:"unit"`
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
) ([]rec, error) {
	var sqlResult []gormRec
	db := s.db
	if tag != "" {
		db = db.Where(datatypes.JSONQuery("tags").HasKey(tag))
	}
	err := db.Order("dis").Find(&sqlResult).Error
	if err != nil {
		return []rec{}, err
	}

	result := []rec{}
	for _, sqlRow := range sqlResult {
		result = append(result, rec(sqlRow))
	}
	return result, nil
}

func (s gormRecStore) readRec(
	id uuid.UUID,
) (*rec, error) {
	var gormRec gormRec
	err := s.db.First(&gormRec, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	rec := rec(gormRec)
	return &rec, nil
}

func (s gormRecStore) createRec(
	rec rec,
) error {
	gormRec := gormRec(rec)
	return s.db.Create(&gormRec).Error
}

func (s gormRecStore) updateRec(
	id uuid.UUID,
	rec rec,
) error {
	var gormRec gormRec
	err := s.db.First(&gormRec, id).Error
	if err != nil {
		return err
	}

	if rec.Dis != nil {
		gormRec.Dis = rec.Dis
	}
	if rec.Unit != nil {
		gormRec.Unit = rec.Unit
	}

	return s.db.Save(&gormRec).Error
}

func (s gormRecStore) deleteRec(id uuid.UUID) error {
	return s.db.Delete(&gormRec{}, id).Error
}

type gormRec struct {
	ID   uuid.UUID         `gorm:"column:id;type:uuid;primaryKey:rec_pkey"`
	Tags datatypes.JSONMap `gorm:"type:json"`
	Dis  *string
	Unit *string
}

func (r gormRec) TableName() string {
	return "rec"
}
