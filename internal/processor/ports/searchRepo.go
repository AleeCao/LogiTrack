package ports

import (
	"context"

	"github.com/AleeCao/LogiTrack/internal/domain"
)

type SearchRepository interface {
	UpdateTruckLocation(ctx context.Context, lcn *domain.Location) error
}
