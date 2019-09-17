package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const RESULT_PATTERN string = "-=result=-"
const ERROR_PATTERN string = "-=error=-"

// Executable hello world
type Executable struct {
	binary            string
	arguments         []string
	stdin             bufio.Scanner
	stdout            bufio.Scanner
	stderr            bufio.Scanner
	timeout           time.Duration
	result            Result
	heartbeatDuration time.Duration
}

func (executable *Executable) Execute(handler MessageHandler) {
	executable.execute(handler)
}

func (executable *Executable) Result() Result {
	return executable.result
}

func (executable *Executable) execute(handler MessageHandler) {
	handler.initialize()
	if handler.receive() {
		if executable.heartbeatDuration.String() != "0s" {
			ticker := time.NewTicker(executable.heartbeatDuration)

			go func() {
				for _ = range ticker.C {
					handler.heartbeat()
				}
			}()

			defer func() {
				ticker.Stop()
			}()
		}
		executable.executableTimeoutHelper(handler)
	}
}

func (executable *Executable) executableTimeoutHelper(handler MessageHandler) {
	errorChannel := make(chan error)
	responseChannel := make(chan *string)
	go func() {
		if response, err := executable.executionHelper(handler); err != nil {
			errorChannel <- err
		} else {
			responseChannel <- response
		}
	}()
	select {
	case err := <-errorChannel:
		log.Printf("E: %s %s", executable.binary, err.Error())
		handler.failure(executable.result)
	case response := <-responseChannel:
		log.Printf("I: %s finished successfully", executable.binary)
		handler.success(response)
	case <-time.After(executable.timeout):
		log.Printf("E: %s timed out after %f seconds", executable.binary, executable.timeout.Seconds())
	}

	defer func() {
		close(errorChannel)
		close(responseChannel)
	}()
}

func inputPipe(pipe io.WriteCloser, inputString *string, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		io.WriteString(pipe, *inputString)
		pipe.Close()
		wg.Done()
	}()
}

func outputPipe(pipe io.ReadCloser, annotation string, result chan *string) {
	pipeScanner := bufio.NewScanner(pipe)
	var response string
	var isResponseString bool

	for pipeScanner.Scan() {
		output := pipeScanner.Text()
		log.Printf("%s %s\n", annotation, output)
		if isResponseString {
			response = output
		}
		if strings.Contains(output, RESULT_PATTERN) || strings.Contains(output, ERROR_PATTERN) {
			isResponseString = true
		} else {
			isResponseString = false
		}
	}

	if (response != "") {
		result <- &response
	} else {
		close(result)
	}
}

func (executable *Executable) executionHelper(handler MessageHandler) (*string, error) {
	var exitCode int
	var err error
	var stdinPipe io.WriteCloser
	var stdoutPipe io.ReadCloser
	var stderrPipe io.ReadCloser

	environ := os.Environ()
	environ = append(environ, fmt.Sprintf("TASK_PAYLOAD=%s", *handler.body()))
	environ = append(environ, fmt.Sprintf("TASK_ID=%s", *handler.id()))
	command := exec.Command(executable.binary, executable.arguments...)
	command.Env = environ

	if handler.body() != nil {
		if stdinPipe, err = command.StdinPipe(); err != nil {
			return nil, err
		}
	}
	if stdoutPipe, err = command.StdoutPipe(); err != nil {
		return nil, err
	}
	if stderrPipe, err = command.StderrPipe(); err != nil {
		return nil, err
	}

	if err = command.Start(); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var stdoutResponse *string
	var stderrResponse *string
	stderrChan := make(chan *string)
	stdoutChan := make(chan *string)

	inputPipe(stdinPipe, handler.body(), &wg)
	go outputPipe(stderrPipe, fmt.Sprintf("%s %s", *handler.id(), "ERROR"), stderrChan)
	go outputPipe(stdoutPipe, fmt.Sprintf("%s", *handler.id()), stdoutChan)

	wg.Wait()

	stdoutResponse = <- stdoutChan
	stderrResponse = <- stderrChan

	if err = command.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.Sys().(syscall.WaitStatus).ExitStatus()
			log.Printf("An error occured (%s %d)\n", executable.binary, exitCode)

			var errorData map[string]interface{}
			
			if stderrResponse != nil {
				json.Unmarshal([]byte(*stderrResponse), &errorData)
			}

			executable.result.Exit = strconv.Itoa(exitCode)
			if errorData["error"] != nil {
				executable.result.Error = fmt.Sprintf("%v", errorData["error"])
			} else {
				executable.result.Error = "ExecutionError"
			}
		}
		return nil, err
	}

	return stdoutResponse, nil
}
