package kafka

import (
	"context"
)

// ConsumerHandler defines the interface for Kafka consumer handlers
type ConsumerHandler interface {
	Start(ctx context.Context) error
	Stop() error
}
