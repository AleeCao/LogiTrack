package services

import (
	"context"

	"github.com/AleeCao/LogiTrack/internal/domain"
	"github.com/AleeCao/LogiTrack/internal/ingestion/ports"
)

type IngestionService struct {
	producer ports.EventProducer
}

func NewIngestionService(producer ports.EventProducer) ports.LocationProcessor {
	return &IngestionService{producer}
}

func (s *IngestionService) ProcessLocation(ctx context.Context, lctn *domain.Location) (domain.Decision, error) {
	if err := lctn.Validate(); err != nil {
		return domain.DecisionAbort, err
	}

	if lctn.TruckStatus == "EMERGENCY" {
		err := s.producer.PublishLocation(ctx, lctn)
		return domain.DecisionComeBack, err
	}
	err := s.producer.PublishLocation(ctx, lctn)
	return domain.DecisionContinue, err
}
