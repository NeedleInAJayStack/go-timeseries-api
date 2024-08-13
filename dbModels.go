package main

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type his struct {
	PointId uuid.UUID       `gorm:"column:pointId;type:uuid;primaryKey:pk_his;index:his_pointId_ts_idx"`
	Ts      *time.Time      `gorm:"primaryKey:pk_his;index:his_pointId_ts_idx,sort:desc;index:his_ts_idx,sort:desc"`
	Value   sql.NullFloat64 `gorm:"type:double precision"`
}
