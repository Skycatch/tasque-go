package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/blaines/tasque-go/result"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// AWSEKS handles the EKS integration
type AWSEKS struct {
	DockerImage       string
	KubeConfigPath    string
	heartbeatDuration time.Duration
	Timeout           time.Duration
	RoleArn           string
}

// Execute executes the Worker on EKS
func (r AWSEKS) Execute(handler MessageHandler) {

	// This inits the handler
	handler.initialize()
	// Gets the message
	handler.receive()
	fmt.Printf("Message received: %s \n", *(handler.body()))

	var clientset *kubernetes.Clientset
	if r.KubeConfigPath != "" {
		// We are on CLI mode
		config, err := clientcmd.BuildConfigFromFlags("", r.KubeConfigPath)
		if err != nil {
			panic(err)
		}

		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}

	} else {
		// We are in cluster
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
		clientset, err = kubernetes.NewForConfig(config)
	}

	v1Client := clientset.CoreV1().Pods("default")

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "vw-",
			Annotations: map[string]string{
				"iam.amazonaws.com/role": r.RoleArn,
			},
		},
		Spec: v1.PodSpec{
			RestartPolicy: "Never",
			NodeSelector: map[string]string{
				"beta.kubernetes.io/instance-type": "i3.xlarge",
			},
			Containers: []v1.Container{
				{
					Name:  "volumetric-worker",
					Image: r.DockerImage,
					Env: []v1.EnvVar{
						{Name: "AWS_REGION", Value: "us-west-2"},
						{Name: "ENVIRONMENT", Value: "development"},
						{Name: "SKYAPI_AUDIENCE", Value: "https://api.skycatch.com"},
						{Name: "SKYAPI_AUTH_URL", Value: "https://skycatch.auth0.com/oauth/token"},
						{Name: "SKYAPI_CLIENT_ID", Value: "e3BhQzZgKaGlt2TtmZqq06DJH6OrlxvU"},
						{Name: "SKYAPI_CLIENT_SECRET",
							ValueFrom: &v1.EnvVarSource{
								SecretKeyRef: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{Name: "skyapi-client-secret"},
									Key:                  "clientsecret",
								},
							},
						},
						{Name: "SKYAPI_URL", Value: "https://api.skycatch.com/v1/"},
						{Name: "TASK_PAYLOAD", Value: *(handler.body())},
					},
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU: resource.MustParse("2000m"),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("500m"),
							v1.ResourceMemory: resource.MustParse("10Gi"),
						},
					},
				},
			},
		},
	}

	// Handle heartbeat
	ticker := time.NewTicker(r.heartbeatDuration)
	timeout := time.After(r.Timeout)
	go func() {
		fmt.Println("Setting up heartbeat handler")
		for range ticker.C {
			fmt.Println("Heartbeat ping")
			handler.heartbeat()
		}
	}()
	fmt.Println("Executing volumetric-worker pod")
	executedPod, err := v1Client.Create(&pod)
	if err != nil {
		handler.failure(result.Result{Error: err.Error(), Exit: fmt.Sprintf("VW %s failed", executedPod.Name)})
	} else {
		watcher, err := v1Client.Watch(metav1.SingleObject(executedPod.ObjectMeta))
		if err != nil {
			handler.failure(result.Result{Error: err.Error(), Exit: fmt.Sprintf("VW %s Failed", executedPod.Name)})
		}
		channel := watcher.ResultChan()
		for {
			select {
			case event := <-channel:
				fmt.Println("Pod event received")
				switch event.Type {
				case watch.Modified:
					pod := event.Object.(*corev1.Pod)
					fmt.Printf("Handling pod status %s\n", pod.Status.Phase)
					switch pod.Status.Phase {
					case "Failed":
						fallthrough
					case "Unknown":
						handler.failure(result.Result{Error: fmt.Sprintf("Pod %s failed", executedPod.Name), Exit: fmt.Sprintf("VW %s Failed", executedPod.Name)})
						defer os.Exit(1)
						return
					case "Succeeded":
						defer os.Exit(0)
						handler.success()
						return
					}
				}
			case <-timeout:
				log.Printf("[INFO] Instance timeout reached.")
				err := fmt.Errorf("Pod %s timed out after %d seconds", executedPod.Name, r.Timeout)
				handler.failure(result.Result{Error: err.Error(), Exit: err.Error()})
				fmt.Printf(err.Error())
				defer os.Exit(1)
				return
			}
		}
	}
}

// Result gets the result of the execution
func (r AWSEKS) Result() result.Result {
	// TODO David: This needs to be worked out. Store the result instead of in-line return.
	return result.Result{}
}
