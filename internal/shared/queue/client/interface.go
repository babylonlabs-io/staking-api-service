package queueclient

type QueueClientInterface interface {
	StartReceivingMessages()
	StopReceivingMessages()
	IsConnectionHealthy() error
}
