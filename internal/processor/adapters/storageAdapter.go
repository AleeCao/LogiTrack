package adapters

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AleeCao/LogiTrack/internal/domain"
	"github.com/AleeCao/LogiTrack/internal/processor/ports"
)

type StorageRepository struct {
	db *sql.DB
}

func NewStorageRepository(db *sql.DB) ports.StorageRepository {
	return &StorageRepository{db}
}

func (a *StorageRepository) CreateLocationRecord(ctx context.Context, lcn *domain.Location) error {
	_, err := a.db.ExecContext(ctx, `INSERT INTO  location_history(truck_id, latitude, longitude, truck_status, timestamp) VALUES ($1, $2, $3, $4, $5)`, lcn.TruckID, lcn.Latitude, lcn.Longitude, lcn.TruckStatus, lcn.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed creating location record: %w", err)
	}

	return nil
}
