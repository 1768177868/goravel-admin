package redisstream

import "errors"

var (
	ErrServiceProviderNotRegistered = errors.New("redisstream service provider not registered")
	ErrQueueConnectionIsRequired    = errors.New("queue connection is required")
)
