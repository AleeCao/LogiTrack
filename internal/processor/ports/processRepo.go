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
	ProcessLocationRecord(ctx context.Context, locations <-chan *domain.Location, cacheChan chan *domain.Location, storageChan chan *domain.Location, wg *sync.WaitGroup) error
	ProcessLocationSearch(ctx context.Context, locations <-chan *domain.Location, wg *sync.WaitGroup) error
}
