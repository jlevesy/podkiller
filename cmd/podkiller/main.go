package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

var (
	frequency     = flag.Duration("frequency", 10*time.Second, "Frequency to stop a control node pod")
	namespace     = flag.String("namespace", "", "Namespace to use")
	labelSelector = flag.String("label-selector", "", "Label selector to apply")
)

func main() {
	log.Println("Pod killer is starting...")
	rand.Seed(time.Now().Unix())
	flag.Parse()

	ticker := time.NewTicker(*frequency)
	defer ticker.Stop()

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for range ticker.C {
		if err := deleteRandomPod(clientset.CoreV1().Pods(*namespace), *labelSelector); err != nil {
			log.Fatalf("unable to kill a pod: %v", err)
		}
	}

	log.Println("Pod killer is stopping...")
}

func deleteRandomPod(client v1.PodInterface, labelSelector string) error {
	// List all possible pods based on gven label selector.
	pods, err := client.List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return fmt.Errorf("unable to list pods: %v", err)
	}

	if len(pods.Items) == 0 {
		log.Printf("No pods found matching %q, see ya next tick !", labelSelector)
		return nil
	}

	// Pick a random pod.
	pod := pods.Items[rand.Intn(len(pods.Items))]
	log.Printf("About to delete pod %q", pod.Name)

	return client.Delete(pod.Name, nil)
}
