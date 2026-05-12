package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/AleeCao/LogiTrack/internal/domain"
	"github.com/AleeCao/LogiTrack/internal/processor/ports"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	Reader   *kafka.Reader
	PrssrSvc ports.ProcessService
}

func NewConsumer(add string, topic string, prss ports.ProcessService) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{add},
		Topic:       topic,
		GroupID:     "processor-consumer-group",
		StartOffset: kafka.FirstOffset,
		MaxBytes:    10e6,
		MinBytes:    10e3,
	})
	reader.Stats()
	fmt.Printf("Consumer initialized with topic: %s\n", topic)
	return &Consumer{reader, prss}
}

func (a *Consumer) StartConsume(ctx context.Context, chanWorkers chan<- *domain.Location, chanES chan<- *domain.Location) error {
	defer a.Reader.Close()
	fmt.Println("start consuming")
	for {

		// Read from the Broker
		a.Reader.Stats()
		msg, err := a.Reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Context cancelled, stopping consumer")
				return nil
			}
			log.Printf("failed consuming from broker: %v", err)
			return fmt.Errorf("failed consuming from broker: %w", err)
		}

		// Unmarshal value into a pointer
		lcn := &domain.Location{}
		err = json.Unmarshal(msg.Value, lcn)
		if err != nil {
			fmt.Printf("error unmarshaling message value: %v", err)
			continue
		}

		// Publish on both channles
		select {
		case chanWorkers <- lcn:
		case <-ctx.Done():
			log.Println("Context cancelled while publishing on workers channel")
			return ctx.Err()
		}

		select {
		case chanES <- lcn:
		case <-ctx.Done():
			log.Println("Context cancelled while publisshing on ES channel")
			return ctx.Err()
		}
	}
}
