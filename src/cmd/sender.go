package main

type MessageSender interface {
	SendMessage(message []byte) (string, error)
}
