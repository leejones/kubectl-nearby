package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"os"
	"regexp"

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
	// if you want to change override values or bind them to flags, there are methods to help you

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
	formattedOutput, err := columnOutput(podsOutput)
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
		containersReadyCount := 0
		var restartCount int32 = 0
		for _, status := range pod.Status.ContainerStatuses {
			if status.Ready {
				containersReadyCount += 1
			}
			restartCount += status.RestartCount
		}
		// TODO: Show only necessary units (e.g. 25s or or 3h5m, or 3d5h), but add tests first
		age := time.Since(pod.CreationTimestamp.Time).Round(time.Second).String()

		pods = append(pods, podInfo{
			age:                  age,
			containersCount:      len(pod.Status.ContainerStatuses),
			containersReadyCount: containersReadyCount,
			name:                 pod.Name,
			namespace:            pod.Namespace,
			restartCount:         restartCount,
			// TODO: Dig into container status to get things like CrashLoopBackup, Terminating, etc. Otherwise it show "Running" in those states.
			status: string(pod.Status.Phase),
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

func columnOutput(input [][]string) (string, error) {
	columnLengths := []int{}
	columnCount := len(input[0])
	for i := 0; i < columnCount; i++ {
		columnLengths = append(columnLengths, 0)
	}
	output := []string{}
	for _, row := range input {
		for index, item := range row {
			currentColumnLength := columnLengths[index]
			if currentColumnLength < len(item) {
				columnLengths[index] = len(item)
			}
		}
	}
	for _, row := range input {
		outputRow := []string{}
		for index, item := range row {
			columnLength := columnLengths[index]
			outputItem := item
			for len(outputItem) < columnLength {
				outputItem += " "
			}
			outputRow = append(outputRow, outputItem)
		}
		output = append(output, strings.Join(outputRow, "  "))
	}
	return strings.Join(output, "\n"), nil
}
