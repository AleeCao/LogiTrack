package ports

import (
	"context"

	"github.com/AleeCao/LogiTrack/internal/domain"
)

type CacheRepository interface {
	SetLocationRecord(ctx context.Context, lcn *domain.Location) error
	GetLocationRecord(ctx context.Context, truckID string) (*domain.Location, error)
}
