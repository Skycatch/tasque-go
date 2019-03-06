package main

import (
	"fmt"

	"github.com/blaines/tasque-go/result"
	jobsv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// AWSEKS handles the EKS integration
type AWSEKS struct {
	DockerImage    string
	KubeConfigPath string
}

// Execute executes the Worker on EKS
func (r AWSEKS) Execute(handler MessageHandler) {

	// This inits the handler
	handler.initialize()
	// Gets the message
	handler.receive()

	// We are in cluster
	var clientset *kubernetes.Clientset
	if r.KubeConfigPath != "" {
		config, err := clientcmd.BuildConfigFromFlags("", r.KubeConfigPath)
		if err != nil {
			panic(err)
		}

		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}

	} else {
		// We are on CLI mode
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
		clientset, err = kubernetes.NewForConfig(config)
	}

	batchClient := clientset.BatchV1().Jobs("default")
	test := jobsv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "vw-",
		},
		Spec: jobsv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "vwpod-",
					Labels: map[string]string{
						"app": "volumetric-worker",
					},
				},
				Spec: v1.PodSpec{
					RestartPolicy: "Never",
					Containers: []v1.Container{
						{
							Name:  "volumetric-worker",
							Image: r.DockerImage,
						},
					},
				},
			},
		},
	}

	executedJob, err := batchClient.Create(&test)
	if err != nil {
		handler.failure(result.Result{Error: err.Error(), Exit: fmt.Sprintf("Job %s failed", executedJob.Name)})
	} else {
		// TODO David: We need to monitor the job till it finishes. It was launched into the cluster but can be long running
		handler.success()
	}
}

// Result gets the result of the execution
func (r AWSEKS) Result() result.Result {
	// TODO David: This needs to be worked out. Store the result instead of in-line return.
	return result.Result{}
}
