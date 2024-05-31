package main

// The MessageSender interface is about delivering data to external services.
// Currently, we send messages to ReportStream or to a local-only mock service for testing.
// Local dev can use either local ReportStream or the mock service
// TODO - have PR env use mock service - any other envs?
type MessageSender interface {
	SendMessage(message []byte) (string, error)
}
