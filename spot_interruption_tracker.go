package main

import (
	"time"
	"net/http"
	"log"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

const actionStop string = "stop"
const actionTerminate string = "terminate"
const url string = "http://169.254.169.254/latest/meta-data/spot/instance-action"

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type SpotInterruptionTracker struct {
	url string
	interval time.Duration
	ticker *time.Ticker
	client HttpClient
}

type InstanceAction struct {
	Action              string `json:"action"`
	Time                string `json:"time"`
}

func NewTracker() *SpotInterruptionTracker {
	return &SpotInterruptionTracker{
		client: &http.Client{Timeout: 5 * time.Second},
		interval: 3 * time.Second,
		url: url,
	}
}

func (tracker *SpotInterruptionTracker) Track(result chan bool) {
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
	instanceAction := tracker.getInstanceAction()
	return instanceAction == actionStop || instanceAction == actionTerminate
}

func (tracker *SpotInterruptionTracker) getInstanceAction() string {
	req, err := http.NewRequest("GET", tracker.url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	resp, err := tracker.client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	if (resp.StatusCode == 404) {
		return "Not found"
	}

	if resp.StatusCode == 200 {
		instanceAction := &InstanceAction{}
		body, _ := ioutil.ReadAll(resp.Body)
		if err := json.Unmarshal(body, instanceAction); err != nil {
			fmt.Println(string(body))
			panic(err)
		}
		return instanceAction.Action
	}
	panic("Spot instance metadata service did not return 200")
}

// func main() {
// 	tracker := NewTracker()
// 	tracker.url = "http://localhost:8080"
// 	result := make(chan bool)
// 	go tracker.Track(result)

// 	time.Sleep(10 * time.Second)
// 	tracker.Untrack()
// 	// <- result
// }