package domain

import (
	"fmt"
	"time"
)

type Location struct {
	TruckID     string    `redis:"truck_ID"`
	Longitude   float64   `redis:"longitude"`
	Latitude    float64   `redis:"latitude"`
	UpdatedAt   time.Time `redis:"updated_at"`
	TruckStatus string    `redis:"truck_status"`
}

func (l *Location) Validate() error {
	if l.TruckID == "" {
		return fmt.Errorf("truck id empty")
	}
	if l.Latitude < -90 || l.Latitude > 90 {
		return fmt.Errorf("invalid latitude")
	}
	if l.Longitude < -180 || l.Longitude > 180 {
		return fmt.Errorf("invalid longitude")
	}
	return nil
}

type Decision int

const (
	DecisionContinue Decision = iota
	DecisionAbort
	DecisionComeBack
)
