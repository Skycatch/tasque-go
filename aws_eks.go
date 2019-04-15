package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"

	"k8s.io/apimachinery/pkg/api/resource"

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

	batchClient := clientset.BatchV1().Jobs("default")

	job := jobsv1.Job{
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
			},
		},
	}

	executedJob, err := batchClient.Create(&job)
	if err != nil {
		handler.failure(result.Result{Error: err.Error(), Exit: fmt.Sprintf("Job %s failed", executedJob.Name)})
	} else {
		// TODO David: We need to monitor the job till it finishes. It was launched into the cluster but can be long running
		spew.Dump(executedJob)
		handler.success()
	}
}

// Result gets the result of the execution
func (r AWSEKS) Result() result.Result {
	// TODO David: This needs to be worked out. Store the result instead of in-line return.
	return result.Result{}
}
