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

func (a *Consumer) StartConsume(ctx context.Context, chn chan domain.Location) error {
	defer a.Reader.Close()
	fmt.Println("start consuming")
	for {
		a.Reader.Stats()
		msg, err := a.Reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("failed consuming from broker: %v", err)
			return fmt.Errorf("failed consuming from broker: %w", err)
		}
		fmt.Printf("Received message: %s\n", string(msg.Value))
		var lcn domain.Location
		err = json.Unmarshal(msg.Value, &lcn)
		if err != nil {
			fmt.Printf("error getting values: %v", err)
		}
		chn <- lcn
	}
}
