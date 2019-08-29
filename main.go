package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
)

// Tasque hello world
type Tasque struct {
	Handler    MessageHandler
	Executable ExecutableInterface
}

func main() {
	var taskDefinition *string
	var overridePayloadKey *string
	var dockerPayloadKey string
	var overrideContainerName *string
	var dockerEndpointPath string
	var deployMethod *string

	isDocker := os.Getenv("DOCKER")
	if isDocker != "" {
		log.Println("Docker mode")
		// Docker Mode
		tasque := Tasque{}
		// DEPLOY_METHOD:  Curerntly it's ECS by default can be switched to DOCKER
		deployMethod = aws.String(os.Getenv("DEPLOY_METHOD"))
		if *deployMethod == "" {
			*deployMethod = "ECS"
		}

		switch strings.ToUpper(*deployMethod) {
		case "DOCKER":
			// DOCKER_CONTAINER_NAME
			overrideContainerName = aws.String(os.Getenv("DOCKER_CONTAINER_NAME"))
			if *overrideContainerName == "" {
				panic("Environment variable DOCKER_CONTAINER_NAME not set")
			}
			// DOCKER_TASK_DEFINITION
			taskDefinition = aws.String(os.Getenv("DOCKER_TASK_DEFINITION"))
			if *taskDefinition == "" {
				panic("Environment variable DOCKER_TASK_DEFINITION not set")
			}

			// DOCKER_ENDPOINT
			dockerEndpointPath = os.Getenv("DOCKER_ENDPOINT")
			if dockerEndpointPath == "" {
				dockerEndpointPath = "unix:///var/run/docker.sock"
			}
			// OVERRIDE_PAYLOAD_KEY
			dockerPayloadKey = os.Getenv("TASK_PAYLOAD")

			overrideTaskDefinition := DockerTaskDefinition{}
			json.Unmarshal([]byte(*taskDefinition), &overrideTaskDefinition)

			d := &AWSDOCKER{
				containerName:        *overrideContainerName,
				timeout:              getTimeout(),
				containerArgs:        dockerPayloadKey,
				dockerTaskDefinition: overrideTaskDefinition,
			}
			d.connect(dockerEndpointPath)
			tasque.Executable = d
			tasque.runWithTimeout()
		case "EKS":
			kubeConfigPath := os.Getenv("KUBE_CONFIG_PATH")
			dockerImage := os.Getenv("VW_DOCKER_IMAGE")
			if dockerImage == "" {
				panic("Environment variable VW_DOCKER_IMAGE not set")
			}
			roleArn := os.Getenv("VW_ROLE_ARN")
			if roleArn == "" {
				panic("Environment variable VW_ROLE_ARN not set")
			}
			tasque.Executable = &AWSEKS{
				DockerImage:       dockerImage,
				KubeConfigPath:    kubeConfigPath,
				heartbeatDuration: time.Second * 60,
				Timeout:           getTimeout(),
				RoleArn:           roleArn,
			}

			tasque.runWithTimeout()
		case "ECS":
			// ECS_TASK_DEFINITION
			taskDefinition = aws.String(os.Getenv("ECS_TASK_DEFINITION"))
			if *taskDefinition == "" {
				panic("Environment variable ECS_TASK_DEFINITION not set")
			}
			// ECS_CONTAINER_NAME
			overrideContainerName = aws.String(os.Getenv("ECS_CONTAINER_NAME"))
			if *overrideContainerName == "" {
				panic("Environment variable ECS_CONTAINER_NAME not set")
			}
			// DOCKER_ENDPOINT
			dockerEndpointPath = os.Getenv("DOCKER_ENDPOINT")
			if dockerEndpointPath == "" {
				dockerEndpointPath = "unix:///var/run/docker.sock"
			}
			// OVERRIDE_PAYLOAD_KEY
			overridePayloadKey = aws.String("TASK_PAYLOAD")
			// DEPLOY_METHOD:  Curerntly it's ECS by default can be switched to DOCKER
			d := &Docker{}
			d.connect(dockerEndpointPath)
			tasque.Executable = &AWSECS{
				docker:                d,
				ecsTaskDefinition:     taskDefinition,
				overrideContainerName: overrideContainerName,
				overridePayloadKey:    overridePayloadKey,
				timeout:               getTimeout(),
			}
			tasque.runWithTimeout()
		default:
			log.Panicf("Unknown or no deployment method provided: %s", *deployMethod)
		}
	} else {
		// CLI Mode
		arguments := os.Args[1:]
		fmt.Printf("%+v", arguments)
		if len(os.Args) > 1 {
			tasque := Tasque{}
			tasque.Executable = &Executable{
				binary:            arguments[0],
				arguments:         arguments[1:],
				timeout:           getTimeout(),
				heartbeatDuration: getHeartbeatTime(),
			}
			tasque.runWithTimeout()
		} else {
			log.Println("Expecting tasque to be run with an application")
			log.Println("Usage: tasque npm start")
		}
	}
}

func (tasque *Tasque) getHandler() {
	var handler MessageHandler
	taskPayload := os.Getenv("TASK_PAYLOAD")
	taskQueueURL := os.Getenv("TASK_QUEUE_URL")
	activityARN := os.Getenv("TASK_ACTIVITY_ARN")
	taskToken := os.Getenv("TASK_TOKEN")

	if taskToken != "" {
		handler = &TokenHandler{
			taskToken: taskToken,
			region: os.Getenv("AWS_REGION"),
		}
	} else if taskPayload != "" {
		handler = &ENVHandler{}
	} else if taskQueueURL != "" {
		handler = &SQSHandler{}
	} else if activityARN != "" {
		handler = &SFNHandler{
			activityARN: activityARN,
		}
	} else {
		panic("No handler")
	}
	tasque.Handler = handler
}

func (tasque *Tasque) runWithTimeout() {
	tasque.getHandler()
	// Commented code is for potential future "daemon"
	// var wg sync.WaitGroup
	// for i := 0; i < 5; i++ {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		for i := 0; i < 5; i++ {
	tasque.Executable.Execute(tasque.Handler)
	// 		}
	// 	}()
	// }
	// wg.Wait()
}

func getTimeout() time.Duration {
	taskTimeout := os.Getenv("TASK_TIMEOUT")
	if taskTimeout == "" {
		log.Println("Default timeout: 30s")
		timeout, _ := time.ParseDuration("30s")
		return timeout
	}
	timeout, err := time.ParseDuration(taskTimeout)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
		return time.Duration(0)
	}
	return timeout
}

func getHeartbeatTime() time.Duration {
	taskTimeout := os.Getenv("TASK_HEARTBEAT")
	if taskTimeout == "" {
		taskTimeout = "30s"
	}
	timeout, err := time.ParseDuration(taskTimeout)
	log.Println("Heartbeat: ", timeout)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
		return time.Duration(0)
	}
	return timeout
}
