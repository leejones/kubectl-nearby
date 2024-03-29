package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"os"
	"regexp"

	"github.com/leejones/kubectl-nearby/pkg/output"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type podsCLI struct {
	allNamespaces bool
	clientset     kubernetes.Clientset
	flags         *flag.FlagSet
	kubeconfig    string
	namespace     string
	podName       string
}

type podInfo struct {
	age                  string
	containersCount      int
	containersReadyCount int
	name                 string
	namespace            string
	restartCount         int32
	status               string
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

	podsCLI.flags = flag.NewFlagSet("kubectl-nearby pods", flag.ContinueOnError)

	podsCLI.flags.Usage = func() {
		fmt.Fprintf(podsCLI.flags.Output(), "List pods on the same node.\n\nUSAGE\n\n  %s pods POD [OPTIONS]\n\nOPTIONS\n\n", os.Args[0])
		podsCLI.flags.PrintDefaults()
	}
	podsCLI.flags.SetOutput(ioutil.Discard)

	var allNamespaces *bool
	allNamespaces = podsCLI.flags.Bool("all-namespaces", false, "Show colocated pods from all namespaces")

	var kubeconfig *string
	kubeconfig = podsCLI.flags.String("kubeconfig", "", fmt.Sprintf("(optional) An absolute path to the kubeconfig file (defaults to the value of KUBECONFIG from the ENV if set or the file %s if present)", clientcmd.RecommendedHomeFile))

	var namespace *string
	namespace = podsCLI.flags.String("namespace", "", "Namespace where the pod is located (defaults to namespace set in kubeconfig if set, otherwise 'default'")

	err = podsCLI.flags.Parse(remainingArgs)
	if err == flag.ErrHelp {
		return &podsCLI, &helpRequestedError{}
	} else if err != nil {
		return &podsCLI, err
	}

	podsCLI.allNamespaces = *allNamespaces
	podsCLI.kubeconfig = *kubeconfig

	// TODO: extract kubeconfig and clientset logic to separate function(s)
	// clientcmd example: https://pkg.go.dev/k8s.io/client-go/tools/clientcmd#pkg-overview

	var loadingRules *clientcmd.ClientConfigLoadingRules
	if podsCLI.kubeconfig == "" {
		// Look in the standard places
		loadingRules = clientcmd.NewDefaultClientConfigLoadingRules()
	} else {
		// Load from given kubeconfig
		loadingRules = &clientcmd.ClientConfigLoadingRules{
			Precedence: []string{podsCLI.kubeconfig},
		}
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	if *namespace == "" {
		podsCLI.namespace, _, err = kubeConfig.Namespace()
		if err != nil {
			return &podsCLI, fmt.Errorf("ERROR: Failed to get namespace from kubeconfig: %v", err)
		}
	} else {
		podsCLI.namespace = *namespace
	}

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return &podsCLI, fmt.Errorf("ERROR: Could not initialize Kubernetes client: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	podsCLI.clientset = *clientset

	return &podsCLI, nil
}

func (podsCLI *podsCLI) execute() error {
	pods, err := podsCLI.fetchPods()
	if err != nil {
		return fmt.Errorf("ERROR: Could not get pods: %v", err)
	}

	podsOutput := [][]string{
		{"NAMESPACE", "NAME", "READY", "STATUS", "RESTARTS", "AGE"},
	}
	for _, pod := range pods {
		containersReady := fmt.Sprintf("%v/%v", pod.containersReadyCount, pod.containersCount)
		podsOutput = append(podsOutput, []string{
			pod.namespace, pod.name, containersReady, pod.status, strconv.FormatInt(int64(pod.restartCount), 10), pod.age,
		})
	}
	formattedOutput, err := output.Columns(podsOutput)
	if err != nil {
		return fmt.Errorf("ERROR: There was an error formatting the output: %v", err)
	}
	fmt.Println(formattedOutput)
	return nil
}

func (podsCLI podsCLI) fetchPods() ([]podInfo, error) {
	var namespaceForList string
	if podsCLI.allNamespaces {
		namespaceForList = ""
	} else {
		namespaceForList = podsCLI.namespace
	}

	podDetails, err := podsCLI.clientset.CoreV1().Pods(podsCLI.namespace).Get(context.TODO(), podsCLI.podName, metav1.GetOptions{})
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}

	// TODO: Should something special happen for unscheduled pods (e.g. status: Pending)?
	// If a pending pod is given, it has no node (it's unscheduled). The search will return
	// all other pods in the same state.
	listOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%v", podDetails.Spec.NodeName),
	}
	podsForNode, err := podsCLI.clientset.CoreV1().Pods(namespaceForList).List(context.TODO(), listOptions)
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}

	var pods []podInfo
	for _, pod := range podsForNode.Items {
		containersReadyCount := 0
		var restartCount int32 = 0
		for _, status := range pod.Status.ContainerStatuses {
			if status.Ready {
				containersReadyCount += 1
			}
			restartCount += status.RestartCount
		}

		age := output.Age(time.Since(pod.CreationTimestamp.Time))

		status := podStatusOutput(pod.Status)

		pods = append(pods, podInfo{
			age:                  age,
			containersCount:      len(pod.Status.ContainerStatuses),
			containersReadyCount: containersReadyCount,
			name:                 pod.Name,
			namespace:            pod.Namespace,
			restartCount:         restartCount,
			status:               status,
		})
	}

	return pods, nil
}

// By default, the flag package shows usage on CLI errors. This
// is a bit noisy and makes the error less obvious. This function
// allows us to disable usage output by default and enable it only
// in specific cases (e.g. --help)
func (podsClI *podsCLI) printUsage() {
	podsClI.flags.SetOutput(os.Stderr)
	podsClI.flags.Usage()
	podsClI.flags.SetOutput(ioutil.Discard)
}

func podStatusOutput(podStatus v1.PodStatus) string {
	output := string(podStatus.Phase)
	if output != "Pending" {
		for _, status := range podStatus.ContainerStatuses {
			if status.State.Waiting != nil {
				return status.State.Waiting.Reason
			} else if status.State.Running != nil {
				output = "Running"
			} else if status.State.Terminated != nil {
				return status.State.Terminated.Reason
			}
		}
	}
	return output
}
