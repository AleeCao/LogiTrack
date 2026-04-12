package adapters

import (
	"context"
	"fmt"

	"github.com/AleeCao/LogiTrack/internal/domain"
	"github.com/AleeCao/LogiTrack/internal/processor/ports"
	"github.com/redis/go-redis/v9"
)

type CacheRepository struct {
	srv *redis.Client
}

func NewCacheRepository(srv *redis.Client) ports.CacheRepository {
	return &CacheRepository{srv}
}

func (a *CacheRepository) SetLocationRecord(ctx context.Context, lcn *domain.Location) error {
	if err := a.srv.HSet(ctx, lcn.TruckID, lcn).Err(); err != nil {
		return fmt.Errorf("failed to set location record: %w", err)
	}
	return nil
}

func (a *CacheRepository) GetLocationRecord(ctx context.Context, truckID string) (*domain.Location, error) {
	data := a.srv.HGetAll(ctx, truckID)

	// Check if the key exists in Redis first
	if data.Err() != nil {
		// Key doesn't exist (first time for this truck) - this is NOT an error
		if data.Err() == redis.Nil {
			return nil, nil
		}
		// Actual error (connection issue, etc.)
		return nil, fmt.Errorf("failed to get location record: %w", data.Err())
	}

	var lcn domain.Location
	err := data.Scan(&lcn)
	if err != nil {
		return nil, fmt.Errorf("failed to scan location record: %w", err)
	}
	if lcn.TruckID == "" {
		return nil, nil
	}

	return &lcn, nil
}
