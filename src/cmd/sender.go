package main

// The MessageSender interface is about delivering data to external services.
// Currently, we send messages to ReportStream or to a mock service for testing.
type MessageSender interface {
	SendMessage(message []byte) (string, error)
}
