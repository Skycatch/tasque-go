package main

type MessageHandler interface {
	id() *string
	body() *string
	initialize()
	receive() bool
	success(*string)
	failure(err Result)
	heartbeat()
}
