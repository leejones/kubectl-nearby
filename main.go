package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/leejones/kubectl-nearby/pkg/nodes"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var buildDate = "unset"
var gitCommit = "unset"
var gitTreeState = "unset"
var version = "unset"

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "ERROR: A command is required.\n")
		printGeneralUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	switch subcommand {
	case "nodes", "node", "no":
		nodesCLI := nodes.NodesCLI{}
		err := nodesCLI.Execute(os.Args[2:], os.Stdout)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
	case "pods", "pod", "po":
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
	case "--version", "--v":
		printVersion()
		os.Exit(0)
	default:
		if subcommand == "" || helpRequested(os.Args) {
			printGeneralUsage()
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: Invalid command or options: %v\n", strings.Join(os.Args[1:], " "))
		}
	}
}

func helpRequested(args [](string)) bool {
	helpMatcher := regexp.MustCompile(`(--help|-h)`)
	for _, arg := range args {
		if helpMatcher.Match([]byte(arg)) {
			return true
		}
	}
	return false
}

func printGeneralUsage() {
	generalUsage := `kubectl-nearby finds nearby pods or nodes.

Commands:
  nodes NODE     List nodes in the same zone as NODE.
  pods POD       List pods on the same node as POD.

Use "kubectl-nearby COMMAND --help" for more information about a specific command.

Global options:

  --version, -v  Display the version and build information.
`
	fmt.Fprint(os.Stderr, generalUsage)
	if !flag.Parsed() {
		flag.Parse()
	}
	flag.PrintDefaults()
}

func printVersion() {
	fmt.Println("Version:", version)
	fmt.Println("BuildDate:", buildDate)
	fmt.Println("GitCommit:", gitCommit)
	fmt.Println("GitTreeState:", gitTreeState)
	fmt.Println("GoVersion:", runtime.Version())
}
