package ports

import (
	"context"

	"github.com/AleeCao/LogiTrack/internal/domain"
)

type LocationProcessor interface {
	ProcessLocation(ctx context.Context, lctn *domain.Location) (domain.Decision, error)
}

type EventProducer interface {
	PublishLocation(ctx context.Context, lctn *domain.Location) error
}
