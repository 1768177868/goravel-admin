package rabbitmq

import "errors"

var (
	ErrServiceProviderNotRegistered = errors.New("rabbitmq service provider not registered")
	ErrQueueConnectionIsRequired    = errors.New("queue connection is required")
)
