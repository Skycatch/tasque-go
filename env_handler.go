package main

import (
	"github.com/blaines/tasque-go/result"
	"os"
)

// ENVHandler hello world
type ENVHandler struct {
	messageID, messageBody string
}

func (handler *ENVHandler) id() *string {
	return &handler.messageID
}

func (handler *ENVHandler) body() *string {
	return &handler.messageBody
}

func (handler *ENVHandler) initialize() {}

func (handler *ENVHandler) receive() bool {
	handler.messageID = "local"
	handler.messageBody = os.Getenv("TASK_PAYLOAD")
	return true
}

func (handler *ENVHandler) success(*string)           {}
func (handler *ENVHandler) failure(err result.Result) {}
func (handler *ENVHandler) heartbeat()                {}
