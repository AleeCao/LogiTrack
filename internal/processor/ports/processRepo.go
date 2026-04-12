// Package ports defines the input and output interfaces for the processor service,
// following Hexagonal Architecture principles to decouple business logic from
// infrastructure-specific implementations.
package ports

import (
	"context"

	"github.com/AleeCao/LogiTrack/internal/domain"
)

type ProcessService interface {
	ProcessLocation(ctx context.Context, lcn *domain.Location) error
}
