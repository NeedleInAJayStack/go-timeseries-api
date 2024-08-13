package main

import (
	"time"

	"github.com/google/uuid"
)

type apiHis struct {
	PointId uuid.UUID  `json:"pointId"`
	Ts      *time.Time `json:"ts"`
	Value   *float64   `json:"value"`
}
