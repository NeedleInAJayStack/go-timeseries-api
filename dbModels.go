package main

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type his struct {
	PointId uuid.UUID
	Ts      *time.Time
	Value   sql.NullFloat64
}
