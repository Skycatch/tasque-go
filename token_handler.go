package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sfn"
)

type TokenHandler struct {
	client      sfn.SFN
	taskToken   string
	messageBody string
	region      string
}

func (handler *TokenHandler) id() *string {
	token := handler.taskToken[0:32]
	return &token
}

func (handler *TokenHandler) body() *string {
	return &handler.messageBody
}

func (handler *TokenHandler) initialize() {
	log.Printf("Configuring handler. token:%s", handler.taskToken)
	sess, err := session.NewSession(&aws.Config{Region: aws.String(handler.region)})
	if err != nil {
		fmt.Println("failed to create session,", err)
		panic("failed to create session")
	}
	handler.client = *sfn.New(sess)
}

func (handler *TokenHandler) receive() bool {
	handler.messageBody = os.Getenv("TASK_PAYLOAD")
	return true
}

func (handler *TokenHandler) success(result *string) {
	sendTaskSuccessParams := &sfn.SendTaskSuccessInput{
		Output:    aws.String(handler.messageBody),
		TaskToken: aws.String(handler.taskToken),
	}

	if result != nil && *result != "" {
		sendTaskSuccessParams.Output = aws.String(*result)
	}

	_, sendMessageError := handler.client.SendTaskSuccess(sendTaskSuccessParams)

	if sendMessageError != nil {
		log.Printf("Couldn't send task success %+v", sendMessageError)
	}
}

func (handler *TokenHandler) failure(err Result) {
	sendTaskFailureParams := &sfn.SendTaskFailureInput{
		TaskToken: aws.String(handler.taskToken),
		Error:     aws.String(err.Error),
		Cause:     aws.String(err.Message()),
	}
	_, deleteMessageError := handler.client.SendTaskFailure(sendTaskFailureParams)

	if deleteMessageError != nil {
		log.Printf("Couldn't send task failure %+v", deleteMessageError)
		return
	}
}

func (handler *TokenHandler) heartbeat() {
	log.Print("Sending heartbeat")
	sendTaskHeartbeatParams := &sfn.SendTaskHeartbeatInput{
		TaskToken: aws.String(handler.taskToken),
	}
	_, deleteMessageError := handler.client.SendTaskHeartbeat(sendTaskHeartbeatParams)

	if deleteMessageError != nil {
		return
	}
}
