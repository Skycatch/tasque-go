package main

import (
	"testing"
)

func TestEnvHandler_id(t *testing.T) {
	got := ENVHandler{"1", "2", "3"}
	want_id := "1"

	if want_id != *got.id() {
		t.Errorf("want message id %s but got %s", want_id, got)
	}
}

func TestEnvHandler_messageBody(t *testing.T) {
	got := ENVHandler{"1", "2", "3"}
	want_message_body := "2"

	if want_message_body != *got.body() {
		t.Errorf("want message id %s but got %s", want_message_body, got)
	}
}

func TestEnvHandler_receive(t *testing.T) {
	got := ENVHandler{"1", "2", "3"}

	if got.receive() == true {
		if got.messageID != "local" {
			t.Errorf("receive messageID %s should have received 'local'", got.messageBody)
		}
		if got.messageBody != got.localPayload {
			t.Errorf("messageBody and localPayload should be identical.")

		}
	} else {
		t.Errorf("receive returned")
	}
}
