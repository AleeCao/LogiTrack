package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/AleeCao/LogiTrack/internal/domain"
	"github.com/AleeCao/LogiTrack/internal/ingestion/ports"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	broker *kafka.Writer
}

func NewProducer(add string, topic string) ports.EventProducer {
	broker := &kafka.Writer{
		Addr:     kafka.TCP(add),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	fmt.Printf("Producer initialized with topic: %s\n", topic)
	return &Producer{broker}
}

func (p *Producer) PublishLocation(ctx context.Context, loc *domain.Location) error {
	jsonLoc, err := json.Marshal(loc)
	if err != nil {
		return fmt.Errorf("failed to marshall location: %w", err)
	}

	err = p.broker.WriteMessages(ctx, kafka.Message{
		Key:   []byte(loc.TruckID),
		Value: []byte(jsonLoc),
	})
	if err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}
	log.Println("Location published: ", loc)
	return nil
}
