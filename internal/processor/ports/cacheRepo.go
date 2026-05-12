package ports

import (
	"context"

	"github.com/AleeCao/LogiTrack/internal/domain"
)

type CacheRepository interface {
	SendBatch(ctx context.Context, buffer *[]*domain.Location)
	GetLocationRecord(ctx context.Context, truckID string) (*domain.Location, error)
}
