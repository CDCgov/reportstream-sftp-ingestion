package main

// The MessageSender interface is about delivering data to external services.
// Currently, we send messages to ReportStream or to a mock service for testing.
type MessageSender interface {
	//TODO - implement these in local file sender?
	//GetPrivateKey() (*rsa.PrivateKey, error)
	//GenerateJwt() (string, error)
	//GetToken() (string, error)

	SendMessage(message []byte) (string, error)
}
