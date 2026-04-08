package nsq

import "errors"

var (
	ErrServiceProviderNotRegistered = errors.New("nsq service provider not registered")
	ErrQueueConnectionIsRequired    = errors.New("queue connection is required")
)
