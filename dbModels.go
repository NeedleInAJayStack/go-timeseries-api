package main

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type his struct {
	PointId uuid.UUID `gorm:"column:pointId"`
	Ts      *time.Time
	Value   sql.NullFloat64
}
