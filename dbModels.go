package main

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type his struct {
	PointId uuid.UUID       `gorm:"column:pointId;type:uuid;primaryKey;index:his_pointId_ts_idx"`
	Ts      *time.Time      `gorm:"primaryKey:pk_his;index:his_pointId_ts_idx,sort:desc;index:his_ts_idx,sort:desc"`
	Value   sql.NullFloat64 `gorm:"type:double precision"`
}

type rec struct {
	ID   uuid.UUID      `gorm:"column:id;type:uuid;primaryKey:rec_pkey"`
	Tags datatypes.JSON `gorm:"type:json"`
	Dis  sql.NullString
	Unit sql.NullString
}

func (rec) TableName() string {
	return "rec"
}
