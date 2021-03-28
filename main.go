package main

import (
	"flag"
	"fmt"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
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
		podsCLI, err := newPodsCLI(os.Args[2:])
		if err != nil {
			helpRequestedError := &helpRequestedError{}
			if err.Error() == helpRequestedError.Error() {
				podsCLI.printUsage()
				os.Exit(0)
			}
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
		err = podsCLI.execute()
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
  pods POD       List pods on the same node as POD

Use "kubectl-nearby COMMAND --help" for more information about a specific command.
`
	fmt.Fprintf(os.Stderr, generalUsage)
	if !flag.Parsed() {
		flag.Parse()
	}
	flag.PrintDefaults()
}
