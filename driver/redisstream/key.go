package redisstream

import "fmt"

type queueKey struct {
	appName    string
	connection string
}

func newQueueKey(appName, connection string) *queueKey {
	return &queueKey{
		appName:    appName,
		connection: connection,
	}
}

func (k *queueKey) stream(queue string) string {
	return fmt.Sprintf("%s_queues:%s_%s:stream", k.appName, k.connection, queue)
}

func (k *queueKey) delayed(queue string) string {
	return fmt.Sprintf("%s_queues:%s_%s:delayed", k.appName, k.connection, queue)
}
