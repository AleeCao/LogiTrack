package ports

import (
	"context"

	"github.com/AleeCao/LogiTrack/internal/domain"
)

type StorageRepository interface {
	InsertLocationRecord(ctx context.Context, lcn *[]*domain.Location) error
}
