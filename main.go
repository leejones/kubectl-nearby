package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/util/homedir"
)

// USAGE
//
// kubectl nearby pods POD [OPTIONS]
//
// OPTIONS
// 	--namespace
//  --all-namespaces
//
// kubectl nearby nodes NODE

func main() {
	var kubeconfig *string
	podsCmd := flag.NewFlagSet("pods", flag.ExitOnError)
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = podsCmd.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = podsCmd.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	var namespace *string
	namespace = podsCmd.String("namespace", "default", "Namespace where the pod is located")

	var allNamespaces *bool
	allNamespaces = podsCmd.Bool("all-namespaces", false, "Show colocated pods from all namespaces")

	subcommand := os.Args[1]
	switch subcommand {
	case "pods":
		if len(os.Args) < 2 {
			// TODO: print helpful error
			printUsage()
			os.Exit(1)
		}
		podName := os.Args[2]
		podsCmd.Parse(os.Args[3:])
		clientset, err := createClient(kubeconfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Could not initialize Kubernetes client: %v", err)
			os.Exit(1)
		}
		err = pods(clientset, podName, *namespace, *allNamespaces)
	case "nodes":
		fmt.Println("not implemented")
		os.Exit(0)
	default:
		if subcommand != "" {
			fmt.Fprintf(os.Stderr, "ERROR: Invalid subcommand: %v\n", subcommand)
		}
		printUsage()
	}

	os.Exit(0)

	// example code below is from:
	// https://github.com/kubernetes/client-go/blob/master/examples/out-of-cluster-client-configuration/main.go

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	for {
		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		// Examples for error handling:
		// - Use helper functions like e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		namespace := "default"
		pod := "example-xxxxx"
		_, err = clientset.CoreV1().Pods(namespace).Get(pod, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %s in namespace %s: %v\n",
				pod, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
		}

		time.Sleep(10 * time.Second)
	}

}

func printUsage() {
	fmt.Fprintf(os.Stderr, "USAGE\n\n%s pods POD [OPTIONS]\n\n", os.Args[0])
	flag.PrintDefaults()
}

func pods(clientset kubernetes.Clientset, pod string, namespace string, allNamespaces bool) error {
	fmt.Println("namespace: ", namespace)
	fmt.Println("allNamespaces: ", allNamespaces)
	podDetails, err := clientset.CoreV1().Pods(namespace).Get(pod, metav1.GetOptions{})
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}
	// fmt.Println(podDetails)
	listOptions := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%v", podDetails.Spec.NodeName),
	}
	podsForNode, err := clientset.CoreV1().Pods(namespace).List(listOptions)
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}
	for _, pod := range podsForNode.Items {
		fmt.Println(pod.Name, pod.Namespace)
	}
	// fmt.Println(podsForNode)
	return nil
}

func createClient(kubeconfig *string) (kubernetes.Clientset, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return *clientset, nil
}
