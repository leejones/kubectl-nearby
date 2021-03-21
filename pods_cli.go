package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/util/homedir"
)

type podsCLI struct {
	allNamespaces bool
	clientset     kubernetes.Clientset
	kubeconfig    string
	namespace     string
	podName       string
}

type podInfo struct {
	name      string
	namespace string
}

type noArgsError struct{}

func (e *noArgsError) Error() string {
	return fmt.Sprintln("A pod name is required, but none was given")
}

type helpRequestedError struct{}

func (e *helpRequestedError) Error() string {
	return fmt.Sprintln("Help requested")
}

func newPodsCLI(args []string) (*podsCLI, error) {
	podsCLI := podsCLI{}
	if len(args) == 0 {
		return &podsCLI, &noArgsError{}
	}

	matched, err := regexp.MatchString("^-", args[0])
	if err != nil {
		return &podsCLI, fmt.Errorf("Error parsing arguments")
	}
	remainingArgs := args
	if !matched {
		podsCLI.podName = args[0]
		if len(args) > 1 {
			remainingArgs = args[1:]
		} else {
			remainingArgs = []string{}
		}
	}

	podsFlags := flag.NewFlagSet("kubectl-nearby pods", flag.ContinueOnError)

	var allNamespaces *bool
	allNamespaces = podsFlags.Bool("all-namespaces", false, "Show colocated pods from all namespaces")

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = podsFlags.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = podsFlags.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	var namespace *string
	namespace = podsFlags.String("namespace", "default", "Namespace where the pod is located")

	err = podsFlags.Parse(remainingArgs)
	if err == flag.ErrHelp {
		return &podsCLI, &helpRequestedError{}
	} else if err != nil {
		return &podsCLI, err
	}

	podsCLI.allNamespaces = *allNamespaces
	podsCLI.kubeconfig = *kubeconfig
	podsCLI.namespace = *namespace

	clientset, err := createClientset(podsCLI.kubeconfig)
	if err != nil {
		return &podsCLI, fmt.Errorf("ERROR: Could not initialize Kubernetes client: %v", err)
	}
	podsCLI.clientset = clientset

	return &podsCLI, nil
}

func (podsCLI *podsCLI) execute() error {
	pods, err := podsCLI.fetchPods()
	if err != nil {
		return fmt.Errorf("ERROR: Could not get pods: %v", err)
	}

	for _, pod := range pods {
		fmt.Println(pod.name, pod.namespace)
	}

	return nil
}

func (podsCLI podsCLI) fetchPods() ([]podInfo, error) {
	var namespaceForList string
	if podsCLI.allNamespaces {
		namespaceForList = ""
	} else {
		namespaceForList = podsCLI.namespace
	}

	podDetails, err := podsCLI.clientset.CoreV1().Pods(podsCLI.namespace).Get(podsCLI.podName, metav1.GetOptions{})
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}

	listOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%v", podDetails.Spec.NodeName),
	}
	podsForNode, err := podsCLI.clientset.CoreV1().Pods(namespaceForList).List(listOptions)
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
