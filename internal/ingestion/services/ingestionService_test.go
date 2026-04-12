package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/AleeCao/LogiTrack/internal/domain"
	"github.com/AleeCao/LogiTrack/internal/ingestion/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventProducer is a mock implementation of ports.EventProducer
type MockEventProducer struct {
	mock.Mock
}

func (m *MockEventProducer) PublishLocation(ctx context.Context, loc *domain.Location) error {
	args := m.Called(ctx, loc)
	return args.Error(0)
}

func TestProcessLocation(t *testing.T) {
	t.Run("Success - Standard Location", func(t *testing.T) {
		// 1. Setup
		mockProducer := new(MockEventProducer)
		service := services.NewIngestionService(mockProducer)
		ctx := context.Background()

		// Arrange
		loc := &domain.Location{
			TruckID:   "truck-123",
			Latitude:  40.7128,
			Longitude: -74.0060,
			UpdatedAt: time.Now(),
		}

		// Expectation: PublishLocation should be called once
		mockProducer.On("PublishLocation", ctx, loc).Return(nil).Once()

		// Act
		decision, err := service.ProcessLocation(ctx, loc)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, domain.DecisionContinue, decision)
		mockProducer.AssertExpectations(t)
	})

	t.Run("Failure - Invalid Location", func(t *testing.T) {
		// 1. Setup
		mockProducer := new(MockEventProducer)
		service := services.NewIngestionService(mockProducer)
		ctx := context.Background()

		// Arrange
		loc := &domain.Location{
			TruckID:   "", // Invalid!
			Latitude:  200.0,
			Longitude: 0,
		}

		// Expectation: PublishLocation should NOT be called
		// (No mock expectation needed, strict mocks will fail if called)

		// Act
		decision, err := service.ProcessLocation(ctx, loc)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, domain.DecisionAbort, decision)
	})

	t.Run("Logic - Emergency Stop", func(t *testing.T) {
		// 1. Setup
		mockProducer := new(MockEventProducer)
		service := services.NewIngestionService(mockProducer)
		ctx := context.Background()

		// Arrange
		loc := &domain.Location{
			TruckID:     "truck-911",
			Latitude:    10.0,
			Longitude:   10.0,
			TruckStatus: "EMERGENCY",
		}

		mockProducer.On("PublishLocation", ctx, loc).Return(nil).Once()

		// Act
		decision, err := service.ProcessLocation(ctx, loc)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, domain.DecisionComeBack, decision)
	})
}
