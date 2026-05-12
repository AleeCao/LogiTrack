package ports

import (
	"context"
	"sync"

	"github.com/AleeCao/LogiTrack/internal/domain"
)

type BatchService interface {
	DbBatcher(ctx context.Context, locations <-chan *domain.Location, wg *sync.WaitGroup)
	CacheBatcher(ctx context.Context, locations <-chan *domain.Location, wg *sync.WaitGroup)
}
