package domain

import "time"

type Carbon struct {
	Intensity int `json:"intensity"`
}

type Slot struct {
	ValidFrom time.Time `json:"valid_from"`
	ValidTo   time.Time `json:"valid_to"`
	Carbon    Carbon    `json:"carbon"`
}
