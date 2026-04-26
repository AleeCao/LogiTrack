// Package ports defines the input and output interfaces for the processor service,
// following Hexagonal Architecture principles to decouple business logic from
// infrastructure-specific implementations.
package ports

import (
	"context"
	"sync"

	"github.com/AleeCao/LogiTrack/internal/domain"
)

type ProcessService interface {
	ProcessLocation(ctx context.Context, locations <-chan domain.Location, wg *sync.WaitGroup) error
}
