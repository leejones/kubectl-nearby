package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"k8s.io/client-go/util/homedir"
)

type podInfo struct {
	name      string
	namespace string
}

type podsRunner struct {
	clientset     kubernetes.Clientset
	namespace     string
	allNamespaces bool
	podName       string
}

func newPodsRunner(args []string) (*podsRunner, error) {
	podsCmd := flag.NewFlagSet("pods", flag.ExitOnError)

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = podsCmd.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = podsCmd.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	var namespace *string
	namespace = podsCmd.String("namespace", "default", "Namespace where the pod is located")

	var allNamespaces *bool
	allNamespaces = podsCmd.Bool("all-namespaces", false, "Show colocated pods from all namespaces")

	podsCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "List pods on the same node.\n\nUSAGE\n\n  %s pods POD [OPTIONS]\n\nOPTIONS\n\n", os.Args[0])
		podsCmd.PrintDefaults()
	}
	podsCmd.Parse(args)

	if len(podsCmd.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "ERROR: A single pod name is required, got: %d\n", len(podsCmd.Args()))
		os.Exit(1)
	}

	podsRunner := podsRunner{}
	podsRunner.podName = args[1]
	podsRunner.namespace = *namespace
	podsRunner.allNamespaces = *allNamespaces

	clientset, err := createClient(kubeconfig)
	if err != nil {
		return &podsRunner, fmt.Errorf("ERROR: Could not initialize Kubernetes client: %v", err)
	}
	podsRunner.clientset = clientset
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Could not initialize Kubernetes client: %v", err)
		os.Exit(1)
	}

	return &podsRunner, nil
}

func (podsRunner *podsRunner) execute() error {
	pods, err := podsRunner.fetchPods()
	if err != nil {
		return fmt.Errorf("ERROR: Could not get pods: %v", err)
	}

	for _, pod := range pods {
		fmt.Println(pod.name, pod.namespace)
	}

	return nil
}

func (podsRunner podsRunner) fetchPods() ([]podInfo, error) {
	var namespaceForList string
	if podsRunner.allNamespaces {
		namespaceForList = ""
	} else {
		namespaceForList = podsRunner.namespace
	}

	podDetails, err := podsRunner.clientset.CoreV1().Pods(podsRunner.namespace).Get(podsRunner.podName, metav1.GetOptions{})
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}

	listOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%v", podDetails.Spec.NodeName),
	}
	podsForNode, err := podsRunner.clientset.CoreV1().Pods(namespaceForList).List(listOptions)
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}

	var pods []podInfo
	for _, pod := range podsForNode.Items {
		pods = append(pods, podInfo{name: pod.Name, namespace: pod.Namespace})
	}

	return pods, nil
}
