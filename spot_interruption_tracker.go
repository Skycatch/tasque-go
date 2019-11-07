package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const actionStop string = "stop"
const actionTerminate string = "terminate"
const url string = "http://169.254.169.254/latest/meta-data/spot/instance-action"

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type SpotInterruptionTracker struct {
	url      string
	interval time.Duration
	ticker   *time.Ticker
	client   HttpClient
}

type InstanceAction struct {
	Action string `json:"action"`
	Time   string `json:"time"`
}

func NewTracker() *SpotInterruptionTracker {
	return &SpotInterruptionTracker{
		client:   &http.Client{Timeout: 5 * time.Second},
		interval: time.Minute,
		url:      url,
	}
}

func (tracker *SpotInterruptionTracker) Track(result chan bool) {
	fmt.Printf("Starting instance interruption tracking with %s interval\n", tracker.interval.String())
	ticker := time.NewTicker(tracker.interval)
	tracker.ticker = ticker

	for _ = range ticker.C {
		if isInterrupting := tracker.isInterrupting(); isInterrupting {
			tracker.Untrack()
			result <- true
			return
		}
	}
}

func (tracker *SpotInterruptionTracker) Untrack() {
	if tracker.ticker != nil {
		tracker.ticker.Stop()
	}
}

func (tracker *SpotInterruptionTracker) isInterrupting() bool {
	instanceAction, err := tracker.getInstanceAction()
	if err != nil {
		log.Printf("%s %s", "Interruption tracker |", err.Error())
		return false
	}
	return instanceAction == actionStop || instanceAction == actionTerminate
}

func (tracker *SpotInterruptionTracker) getInstanceAction() (string, error) {
	req, err := http.NewRequest("GET", tracker.url, nil)
	if err != nil {
		return "", err
	}
	resp, err := tracker.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "Not found", nil
	}

	if resp.StatusCode == 200 {
		instanceAction := &InstanceAction{}
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Interruption tracker | response body: ", string(body))
		if err := json.Unmarshal(body, instanceAction); err != nil {
			fmt.Println(string(body))
			return "", err
		}
		return instanceAction.Action, nil
	}

	return "", fmt.Errorf("Spot instance metadata service returned %d code", resp.StatusCode)
}
