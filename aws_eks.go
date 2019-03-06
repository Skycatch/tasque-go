package main

import (
	"path/filepath"

	"github.com/blaines/tasque-go/result"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// AWSEKS handles the EKS integration
type AWSEKS struct {
	DockerImage string
}

// Execute executes the Worker on EKS
func (r AWSEKS) Execute(handler MessageHandler) {

	// This inits the handler
	handler.initialize()
	// Gets the message
	handler.receive()
	kubeconfig := filepath.Join("/home/david", ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	_, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	//clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	//spew.Dump(err)
	//spew.Dump(pods.ListMeta)
	//batchClient := clientset.BatchV1().Jobs("default")
	//batchv1.Job{}
}

// Result gets the result of the execution
func (r AWSEKS) Result() result.Result {

	return result.Result{}
}
