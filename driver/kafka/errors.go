package kafka

import "errors"

var (
	ErrServiceProviderNotRegistered = errors.New("kafka service provider not registered")
	ErrQueueConnectionIsRequired    = errors.New("queue connection is required")
)
