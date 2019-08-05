package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
	"encoding/json"
	"strconv"

	"github.com/blaines/tasque-go/result"
)

// Executable hello world
type Executable struct {
	binary    string
	arguments []string
	stdin     bufio.Scanner
	stdout    bufio.Scanner
	stderr    bufio.Scanner
	timeout   time.Duration
	result    result.Result
}

func (executable *Executable) Execute(handler MessageHandler) {
	executable.execute(handler)
}

func (executable *Executable) Result() result.Result {
	return executable.result
}

func (executable *Executable) execute(handler MessageHandler) {
	handler.initialize()
	if handler.receive() {
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

func outputPipe(pipe io.ReadCloser, annotation string, wg *sync.WaitGroup, collectResponse bool) *string {
	wg.Add(1)
	pipeScanner := bufio.NewScanner(pipe)
	ch := make(chan string)
	var response string

	go func() {
		for pipeScanner.Scan() {
			output := pipeScanner.Text()
			log.Printf("%s %s\n", annotation, output)
			ch <- output
		}
		close(ch)
		wg.Done()
	}()

	for {
		output, more := <-ch
		if !more {
			return &response
		}
		if collectResponse {
			response += output
		}
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
	inputPipe(stdinPipe, handler.body(), &wg)
	stderrResponse := outputPipe(stderrPipe, fmt.Sprintf("%s %s", *handler.id(), "ERROR"), &wg, handler.returnResult())
	stdoutResponse := outputPipe(stdoutPipe, fmt.Sprintf("%s", *handler.id()), &wg, handler.returnResult())
	wg.Wait()
	if err != nil {
		return nil, err
	}

	if err = command.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.Sys().(syscall.WaitStatus).ExitStatus()
			log.Printf("An error occured (%s %d)\n", executable.binary, exitCode)
			
			var errorData map[string]interface{}
			json.Unmarshal([]byte(*stderrResponse), &errorData)

			executable.result.Exit = strconv.Itoa(exitCode)
			if errorData["error"] != nil {
				executable.result.Error = fmt.Sprintf("%v", errorData["error"])
			}
		}
		return nil, err
	}

	return stdoutResponse, nil
}
