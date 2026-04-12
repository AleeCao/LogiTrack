// Package domain defines the core data structures used by the processor service.
// This includes specialized types for Elasticsearch serialization to ensure
// geospatial data is indexed correctly as geo_points.
package domain

import "time"

type GeoPoint struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type EsLocation struct {
	TruckID     string    `json:"truck_ID"`
	Location    GeoPoint  `json:"location"`
	UpdatedAt   time.Time `json:"updated_at"`
	TruckStatus string    `json:"truck_status"`
}
