package adapters

import (
	"context"
	"fmt"
	"log"

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

func (a *CacheRepository) SendBatch(ctx context.Context, buffer *[]*domain.Location) {
	if len(*buffer) == 0 {
		return
	}

	pipe := a.srv.Pipeline()

	for _, op := range *buffer {
		if err := pipe.HSet(ctx, op.TruckID, op).Err(); err != nil {
			log.Println("failed to set location record: ", err)
		}
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Println("error dispatching: ", err)
	}

	*buffer = nil
}

func (a *CacheRepository) GetLocationRecord(ctx context.Context, truckID string) (*domain.Location, error) {
	data := a.srv.HGetAll(ctx, truckID)

	// Check if the key exists in Redis first
	if data.Err() != nil {
		// Key doesn't exist (first time for this truck)
		if data.Err() == redis.Nil {
			return nil, nil
		}
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
