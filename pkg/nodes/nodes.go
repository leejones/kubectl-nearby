// TODO: describe package
package nodes

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/leejones/kubectl-nearby/pkg/output"
)

type NodesCLI struct {
	Client kubernetes.Interface
}

type ErrNodeNameRequired struct{}

func (err ErrNodeNameRequired) Error() string {
	return "a node name is required"
}

// TODO: describe method
func (n *NodesCLI) Execute(args []string, writer io.Writer) error {
	var nodeName string
	var remainingArgs []string

	if len(args) > 0 {
		matched, err := regexp.MatchString("^-", args[0])
		if err != nil {
			return fmt.Errorf("Error parsing arguments")
		}
		remainingArgs = args
		if !matched {
			nodeName = args[0]
			if len(args) > 1 {
				remainingArgs = args[1:]
			} else {
				remainingArgs = []string{}
			}
		}

	}

	f := flag.NewFlagSet("kubectl nearby nodes", flag.ContinueOnError)
	f.Usage = func() {
		fmt.Fprintf(f.Output(), "List nodes in the same node.\n\nUSAGE\n\n  %s node NODE [OPTIONS]\n\nOPTIONS\n\n", os.Args[0])
		f.PrintDefaults()
	}
	f.SetOutput(ioutil.Discard)

	kubeconfig := f.String("kubeconfig", "", fmt.Sprintf("(optional) An absolute path to the kubeconfig file (defaults to the value of KUBECONFIG from the ENV if set or the file %s if present)", clientcmd.RecommendedHomeFile))

	err := f.Parse(remainingArgs)
	if err == flag.ErrHelp {
		Usage(f, writer)
		return nil
	} else if err != nil {
		return fmt.Errorf("error parsing CLI arguments: %v", err)
	}

	if nodeName == "" {
		return ErrNodeNameRequired{}
	}

	if n.Client == nil {
		n.Client, err = DefaultClient(*kubeconfig)
		if err != nil {
			return fmt.Errorf("error setting up default client: %v", err)
		}
	}

	node, err := n.Client.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable fetch node: %v", err)
	}

	zone, ok := node.Labels["topology.kubernetes.io/zone"]
	if !ok {
		return fmt.Errorf("unable to find label 'topology.kubernetes.io/zone' on node: %v", node.Name)
	}

	nearbyNodes, err := n.Client.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("topology.kubernetes.io/zone=%v", zone),
	})
	if err != nil {
		return fmt.Errorf("unable to fetch nearby nodes: %v", err)
	}
	nodesOutput := [][]string{
		{"NAME", "STATUS", "ROLES", "AGE", "VERSION", "ZONE"},
	}

	for _, node := range nearbyNodes.Items {
		roles := []string{}
		for key, _ := range node.Labels {
			if strings.HasPrefix(key, "node-role.kubernetes.io/") {
				roleParts := strings.Split(key, "/")
				if len(roleParts) == 2 {
					roles = append(roles, roleParts[1])
				}
			}
		}
		var rolesOutput string
		if len(roles) > 0 {
			rolesOutput = strings.Join(roles, ",")
		} else {
			rolesOutput = "<none>"
		}
		zone, ok := node.Labels["topology.kubernetes.io/zone"]
		if !ok {
			zone = "<unknown>"
		}
		age := output.Age(time.Since(node.CreationTimestamp.Time))
		status := "<unknown>"
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" {
				switch condition.Status {
				case v1.ConditionTrue:
					status = "Ready"
				case v1.ConditionFalse:
					status = "NotReady"
				case v1.ConditionUnknown:
					status = "Unknown"
				}
			}
		}
		nodesOutput = append(nodesOutput, []string{
			node.Name, status, rolesOutput, age, node.Status.NodeInfo.KubeletVersion, zone,
		})
	}
	output, err := output.Columns(nodesOutput)
	if err != nil {
		return fmt.Errorf("columized output: %v", err)
	}
	fmt.Fprintln(writer, output)
	return nil
}

func Usage(flags *flag.FlagSet, writer io.Writer) {
	flags.SetOutput(writer)
	flags.Usage()
	flags.SetOutput(ioutil.Discard)
}

func DefaultClient(kubeconfig string) (*kubernetes.Clientset, error) {
	var loadingRules *clientcmd.ClientConfigLoadingRules
	if kubeconfig == "" {
		// Look in the standard places
		loadingRules = clientcmd.NewDefaultClientConfigLoadingRules()
	} else {
		_, err := os.Stat(kubeconfig)
		if os.IsNotExist(err) {
			return &kubernetes.Clientset{}, fmt.Errorf("config file: %v", err)
		}
		// Load from given kubeconfig
		loadingRules = &clientcmd.ClientConfigLoadingRules{
			Precedence: []string{kubeconfig},
		}
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	clientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return &kubernetes.Clientset{}, fmt.Errorf("could not initialize Kubernetes client config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return &kubernetes.Clientset{}, fmt.Errorf("could not create clientset from config: %v", err)
	}
	return clientset, nil
}
