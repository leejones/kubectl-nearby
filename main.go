package main

import (
	"flag"
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
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
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "ERROR: A subcommand is required.\n")
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	switch subcommand {
	case "pods":
		podsRunner, err := newPodsRunner(os.Args)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
		err = podsRunner.execute()
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
	case "nodes":
		fmt.Println("not implemented")
		os.Exit(0)
	default:
		if subcommand != "" {
			fmt.Fprintf(os.Stderr, "ERROR: Invalid subcommand: %v\n", subcommand)
		}
		printUsage()
	}
}

// TODO: move to pods runner
func printUsage() {
	fmt.Fprintf(os.Stderr, "USAGE\n\n%s pods POD [OPTIONS]\n\n", os.Args[0])
	flag.PrintDefaults()
}

// TODO: move to separate file
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
