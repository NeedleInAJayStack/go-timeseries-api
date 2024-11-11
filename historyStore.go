package main

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// historyStore is able to store point historical values
type historyStore interface {
	readHistory(uuid.UUID, *time.Time, *time.Time) ([]apiHisItem, error)
	writeHistory(uuid.UUID, apiHisItem) error
	deleteHistory(uuid.UUID, *time.Time, *time.Time) error
}

type apiHis struct {
	PointId uuid.UUID  `json:"pointId"`
	Ts      *time.Time `json:"ts"`
	Value   *float64   `json:"value"`
}

type apiHisItem struct {
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
) ([]apiHisItem, error) {
	result := []apiHisItem{}

	var sqlResult []his
	query := s.db.Where(&his{PointId: pointId})
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
		result = append(result, apiHisItem{Ts: sqlRow.Ts, Value: sqlRow.Value})
	}
	return result, nil
}

func (s gormHistoryStore) writeHistory(
	pointId uuid.UUID,
	hisItem apiHisItem,
) error {
	his := his{
		PointId: pointId,
		Ts:      hisItem.Ts,
		Value:   hisItem.Value,
	}

	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pointId"}, {Name: "ts"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&his).Error
}

func (s gormHistoryStore) deleteHistory(
	pointId uuid.UUID,
	start *time.Time,
	end *time.Time,
) error {
	var sqlResult []his
	query := s.db.Where(&his{PointId: pointId})
	if start != nil {
		query.Where("ts >= ?", start)
	}
	if end != nil {
		query.Where("ts < ?", end)
	}
	return query.Delete(&sqlResult).Error
}

type his struct {
	PointId uuid.UUID  `gorm:"column:pointId;type:uuid;primaryKey;index:his_pointId_ts_idx"`
	Ts      *time.Time `gorm:"primaryKey:pk_his;index:his_pointId_ts_idx,sort:desc;index:his_ts_idx,sort:desc"`
	Value   *float64   `gorm:"type:double precision"`
}
