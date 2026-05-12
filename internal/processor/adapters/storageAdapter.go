package adapters

import (
	"context"
	"fmt"
	"log"

	"github.com/AleeCao/LogiTrack/internal/domain"
	"github.com/AleeCao/LogiTrack/internal/processor/ports"
	"github.com/jackc/pgx/v5"
)

type StorageRepo struct {
	db *pgx.Conn
}

func NewStorageRepository(db *pgx.Conn) ports.StorageRepository {
	return &StorageRepo{db}
}

func (a *StorageRepo) InsertLocationRecord(ctx context.Context, lcn *[]*domain.Location) error {
	var e *domain.Location
	count, err := a.db.CopyFrom(ctx, pgx.Identifier{"location_history"}, []string{"truck_id", "latitude", "longitude", "truck_status", "timestamp"}, pgx.CopyFromSlice(len(*lcn), func(i int) ([]any, error) {
		e = (*lcn)[i]
		return []any{e.TruckID, e.Latitude, e.Longitude, e.TruckStatus, e.UpdatedAt}, nil
	}))
	if err != nil {
		return fmt.Errorf("failed creating location record: %w", err)
	}
	log.Printf("%d rows inserted", count)
	*lcn = nil
	return nil
}
