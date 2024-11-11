package main

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// historyStore is able to store point historical values
type historyStore interface {
	readHistory(uuid.UUID, *time.Time, *time.Time) ([]hisItem, error)
	writeHistory(uuid.UUID, hisItem) error
	deleteHistory(uuid.UUID, *time.Time, *time.Time) error
}

type hisItem struct {
	Ts    *time.Time `json:"ts"`
	Value *float64   `json:"value"`
}

// gormHistoryStore stores point historical values in a GORM database.
type gormHistoryStore struct {
	db *gorm.DB
}

func newGormHistoryStore(db *gorm.DB) gormHistoryStore {
	return gormHistoryStore{db: db}
}

func (s gormHistoryStore) readHistory(
	pointId uuid.UUID,
	start *time.Time,
	end *time.Time,
) ([]hisItem, error) {
	result := []hisItem{}

	var sqlResult []gormHis
	query := s.db.Where(&gormHis{PointId: pointId})
	if start != nil {
		query.Where("ts >= ?", start)
	}
	if end != nil {
		query.Where("ts < ?", end)
	}
	err := query.Order("ts asc").Find(&sqlResult).Error
	if err != nil {
		return result, err
	}
	for _, sqlRow := range sqlResult {
		result = append(result, hisItem{Ts: sqlRow.Ts, Value: sqlRow.Value})
	}
	return result, nil
}

func (s gormHistoryStore) writeHistory(
	pointId uuid.UUID,
	hisItem hisItem,
) error {
	gormHis := gormHis{
		PointId: pointId,
		Ts:      hisItem.Ts,
		Value:   hisItem.Value,
	}

	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pointId"}, {Name: "ts"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&gormHis).Error
}

func (s gormHistoryStore) deleteHistory(
	pointId uuid.UUID,
	start *time.Time,
	end *time.Time,
) error {
	var sqlResult []gormHis
	query := s.db.Where(&gormHis{PointId: pointId})
	if start != nil {
		query.Where("ts >= ?", start)
	}
	if end != nil {
		query.Where("ts < ?", end)
	}
	return query.Delete(&sqlResult).Error
}

type gormHis struct {
	PointId uuid.UUID  `gorm:"column:pointId;type:uuid;primaryKey;index:his_pointId_ts_idx"`
	Ts      *time.Time `gorm:"primaryKey:pk_his;index:his_pointId_ts_idx,sort:desc;index:his_ts_idx,sort:desc"`
	Value   *float64   `gorm:"type:double precision"`
}

func (gormHis) TableName() string {
	return "his"
}
