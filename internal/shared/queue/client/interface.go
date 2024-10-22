package queueclient

type QueueClient interface {
	StartReceivingMessages()
	StopReceivingMessages()
	IsConnectionHealthy() error
}
