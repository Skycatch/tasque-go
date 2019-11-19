package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
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
	binary                    string
	arguments                 []string
	stdin                     bufio.Scanner
	stdout                    bufio.Scanner
	stderr                    bufio.Scanner
	timeout                   time.Duration
	result                    Result
	heartbeatDuration         time.Duration
	trackInstanceInterruption bool
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

			defer ticker.Stop()
		}

		executable.executableTimeoutHelper(handler)
	}
}

func (executable *Executable) executableTimeoutHelper(handler MessageHandler) {
	errorChannel := make(chan error)
	responseChannel := make(chan *string)
	stopChannel := make(chan string)
	instanceInterruptionChannel := make(chan bool)
	scriptInterruptionChannel := make(chan bool)

	handleSignals(scriptInterruptionChannel)

	if executable.trackInstanceInterruption {
		tracker := NewTracker()
		go tracker.Track(instanceInterruptionChannel)
		defer tracker.Untrack()
	}

	go func() {
		if response, err := executable.executionHelper(handler, stopChannel); err != nil {
			errorChannel <- err
		} else {
			responseChannel <- response
		}
	}()

	go func() {
		select {
		case instanceInterrupted := <-instanceInterruptionChannel:
			if instanceInterrupted {
				fmt.Println("Instance interruption event caught")
				stopChannel <- "InstanceInterruption"
			}
		case scriptInterrupted := <-scriptInterruptionChannel:
			if scriptInterrupted {
				fmt.Println("Script interruption signal caught")
				stopChannel <- "ScriptInterruption"
			}
		case <-time.After(executable.timeout):
			stopChannel <- "Timeout"
			log.Printf("E: %s timed out after %.0f seconds", executable.binary, executable.timeout.Seconds())
		}
	}()

	select {
	case err := <-errorChannel:
		log.Printf("E: %s %s\n", executable.binary, err.Error())
		handler.failure(executable.result)
	case response := <-responseChannel:
		fmt.Printf("I: %s finished successfully\n", executable.binary)
		handler.success(response)
	}

	defer func() {
		close(errorChannel)
		close(responseChannel)
		close(stopChannel)
		close(instanceInterruptionChannel)
		close(scriptInterruptionChannel)
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
		fmt.Printf("%s %s\n", annotation, output)
		if isResponseString {
			response = output
		}
		if strings.Contains(output, RESULT_PATTERN) || strings.Contains(output, ERROR_PATTERN) {
			isResponseString = true
		} else {
			isResponseString = false
		}
	}

	if response != "" {
		result <- &response
	} else {
		close(result)
	}
}

func (executable *Executable) executionHelper(handler MessageHandler, stopChannel chan string) (*string, error) {
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

	var stopChannelError string
	var wg sync.WaitGroup
	var stdoutResponse *string
	var stderrResponse *string
	stderrChan := make(chan *string)
	stdoutChan := make(chan *string)

	inputPipe(stdinPipe, handler.body(), &wg)
	go outputPipe(stderrPipe, fmt.Sprintf("%s %s", *handler.id(), "ERROR"), stderrChan)
	go outputPipe(stdoutPipe, fmt.Sprintf("%s", *handler.id()), stdoutChan)

	wg.Wait()

	go func() {
		select {
		case stdoutResponse = <-stdoutChan:
		case stderrResponse = <-stderrChan:
		case stop := <-stopChannel:
			fmt.Printf("E: %s\n", stop)
			executable.result.Error = stop
			stopChannelError = stop
			// since windows env doesn't support interruption signal, killing process immediately
			if runtime.GOOS != "windows" {
				command.Process.Signal(os.Interrupt)
				fmt.Println("Interrupt signal sent")
				time.Sleep(5 * time.Second)
			}
			fmt.Println("killing worker process")
			command.Process.Kill()
		}
	}()

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
			}
		}
		return nil, err
	}

	if stopChannelError != "" {
		return nil, errors.New(stopChannelError)
	}

	return stdoutResponse, nil
}

func handleSignals(interruptionChannel chan<- bool) {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		<-sigs
		interruptionChannel <- true
	}()
}
