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
		fmt.Fprintf(os.Stderr, "ERROR: A command is required.\n")
		printGeneralUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	switch subcommand {
	case "nodes":
		fmt.Println("not implemented")
		os.Exit(0)
	case "pods":
		podsRunner, err := newPodsRunner(os.Args[2:])
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
		err = podsRunner.execute()
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
	default:
		if subcommand != "" {
			fmt.Fprintf(os.Stderr, "ERROR: Invalid command: %v\n", subcommand)
		}
	}
}

func printGeneralUsage() {
	generalUsage := `kubectl-nearby finds nearby pods or nodes.

Commands:
  nodes NODE     List nodes in the same availability zone as NODE
  podes POD      List pods on the same node as POD

Use "kubectl-nearby COMMAND	--help" for more information about a specific command.
`
	fmt.Fprintf(os.Stderr, generalUsage)
	if !flag.Parsed() {
		flag.Parse()
	}
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
