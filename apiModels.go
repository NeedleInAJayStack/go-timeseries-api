package main

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type apiHis struct {
	PointId uuid.UUID  `json:"pointId"`
	Ts      *time.Time `json:"ts"`
	Value   *float64   `json:"value"`
}

type apiHisItem struct {
	Ts    *time.Time `json:"ts"`
	Value *float64   `json:"value"`
}

type apiRec struct {
	ID   uuid.UUID         `json:"id"`
	Tags datatypes.JSONMap `json:"tags"`
	Dis  *string           `json:"dis"`
	Unit *string           `json:"unit"`
}

type apiCurrentInput struct {
	Value *float64 `json:"value"`
}

type apiCurrent struct {
	Ts    *time.Time `json:"ts"`
	Value *float64   `json:"value"`
}

type clientToken struct {
	Token string `json:"token"`
}
