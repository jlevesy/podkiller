package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
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
	amount        = flag.String("amount", "1", "Amount of pod to stop, all to stop them all")
)

func main() {
	log.Println("Pod killer is starting...")
	rand.Seed(time.Now().Unix())
	flag.Parse()

	var (
		total int
		all   bool
	)

	if *amount == "" {
		log.Fatalf("Must provide an amount of pods to stop")
	}

	total, _ = strconv.Atoi(*amount)
	all = *amount == "all"

	if all {
		log.Println("I'm gonna stop them all")
	} else {
		log.Printf("I'm gonna stop %d pods", total)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	ticker := time.NewTicker(*frequency)
	defer ticker.Stop()

	for range ticker.C {
		if err := deletePods(clientset.CoreV1().Pods(*namespace), *labelSelector, total, all); err != nil {
			log.Fatalf("unable to kill a pod: %v", err)
		}
	}

	log.Println("Pod killer is stopping...")
}

func deletePods(client v1.PodInterface, labelSelector string, total int, all bool) error {
	// List all possible pods based on gven label selector.
	pods, err := client.List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return fmt.Errorf("unable to list pods: %v", err)
	}

	if len(pods.Items) == 0 {
		log.Printf("No pods found matching %q, see ya next tick !", labelSelector)
		return nil
	}

	if len(pods.Items) < total {
		log.Printf("No enough pods to stop, I'll stop them all")
		total = len(pods.Items)
	}

	if all {
		total = len(pods.Items)
	}

	// Shuffle pods.
	rand.Shuffle(len(pods.Items), func(i, j int) { pods.Items[i], pods.Items[j] = pods.Items[j], pods.Items[i] })

	for i := 0; i < total; i++ {
		pod := pods.Items[i]

		log.Printf("Deleting pod %q", pod.Name)
		if err = client.Delete(pod.Name, nil); err != nil {
			return fmt.Errorf("unable to stop pod %q: %v", pod.Name, err)
		}
	}

	return nil
}
