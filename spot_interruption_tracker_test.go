package main

import (
	"testing"
	"net/http"
	"io/ioutil"
	"bytes"
	"time"
	"fmt"
	"strconv"
)

type HttpClientMock struct {
	body string
	status int
}


func (c *HttpClientMock) Do(req *http.Request) (*http.Response, error) {
	return c.createResponse()
}

func (c *HttpClientMock) createResponse() (*http.Response, error) {
	return &http.Response{
		Status:        fmt.Sprintf("%s %s", strconv.Itoa(c.status), "OK"),
		StatusCode:    c.status,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body: ioutil.NopCloser(bytes.NewBufferString(c.body)),
		ContentLength: int64(len(c.body)),
		Header:        make(http.Header, 0),
	}, nil
}

func createTracker(client HttpClient) *SpotInterruptionTracker {
	return &SpotInterruptionTracker{
		client: client,
		interval: (1 * time.Millisecond),
	}
}

func createClient(body string, params ...int) HttpClient {
	status := 200
	if len(params) > 0 {
		status = params[0]
	}
	return &HttpClientMock{
		body: body,
		status: status,
	}
}

func TestReturnsOnStop(t *testing.T) {
	client := createClient(`{"action":"stop","time":"10:10"}`)
	tracker := createTracker(client)
	resultChannel := make(chan bool)
	go tracker.Track(resultChannel)
	select {
	case <-resultChannel:
	case <-time.After(1 * time.Second):
		t.Errorf("Expected to return true on stop action received")
	}
}

func TestReturnsOnTerminate(t *testing.T) {
	client := createClient(`{"action":"terminate","time":"10:10"}`)
	tracker := createTracker(client)
	resultChannel := make(chan bool)
	go tracker.Track(resultChannel)
	select {
	case <-resultChannel:
	case <-time.After(1 * time.Second):
		t.Errorf("Expected to return true on terminate action received")
	}
}

func TestDoesNotReturnOnNonTerminate(t *testing.T) {
	client := createClient(`{"action":"terminat","time":"10:10"}`)
	tracker := createTracker(client)
	resultChannel := make(chan bool)
	go tracker.Track(resultChannel)
	select {
	case <-resultChannel:
		t.Errorf("Expected not to return result if termination action didn't happen")
	case <-time.After(5 * time.Millisecond):
	}
}

func TestDoesNotReturnOn404(t *testing.T) {
	client := createClient(`Not found`, 404)
	tracker := createTracker(client)
	resultChannel := make(chan bool)
	go tracker.Track(resultChannel)
	select {
	case <-resultChannel:
		t.Errorf("Expected not to return result if termination action didn't happen")
	case <-time.After(5 * time.Millisecond):
	}
}
