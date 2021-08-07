package nodes_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"

	"github.com/leejones/kubectl-nearby/pkg/nodes"
)

func TestExecute(t *testing.T) {
	t.Run("with --help", func(t *testing.T) {
		expected := "USAGE"
		args := []string{"--help"}
		writer := bytes.NewBufferString("")
		nodesCLI := nodes.NodesCLI{
			Client: testclient.NewSimpleClientset(),
		}
		err := nodesCLI.Execute(args, writer)
		if err != nil {
			t.Errorf("Unexpected error: %v\n", err)
		}
		output, _ := ioutil.ReadAll(writer)
		if !strings.Contains(string(output), expected) {
			t.Errorf("Expected help output to include: %v, got: \n%v", expected, string(output))
		}
	})

	t.Run("with no node name, it returns an error", func(t *testing.T) {
		writer := bytes.NewBufferString("")
		nodesCLI := nodes.NodesCLI{
			Client: testclient.NewSimpleClientset(),
		}
		err := nodesCLI.Execute([]string{}, writer)
		// Test error's type using type assertion (https://tour.golang.org/methods/15)
		got, ok := err.(nodes.ErrNodeNameRequired)
		if !ok {
			t.Errorf("Expected error type: %T, got: %T\n", got, err)
		}
	})

	t.Run("with a node name, returns a list of nodes in the same zone", func(t *testing.T) {
		writer := bytes.NewBufferString("")

		// Helpful blog posts for faking k8s api in unit tests:
		// https://gianarb.it/blog/unit-testing-kubernetes-client-in-go
		// https://medium.com/the-phi/mocking-the-kubernetes-client-in-go-for-unit-testing-ddae65c4302
		clientset := testclient.NewSimpleClientset(
			&v1.Node{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Node",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-a-1",
					Labels: map[string]string{
						"topology.kubernetes.io/zone": "us-east4-a",
					},
					CreationTimestamp: metav1.NewTime(time.Now().Add(time.Hour * -1)),
				},
				Status: v1.NodeStatus{
					NodeInfo: v1.NodeSystemInfo{
						KubeletVersion: "1.19.10",
					},
					Conditions: []v1.NodeCondition{
						{
							Type:   "Ready",
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			&v1.Node{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Node",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-a-2",
					Labels: map[string]string{
						"topology.kubernetes.io/zone": "us-east4-a",
					},
					CreationTimestamp: metav1.NewTime(time.Now().Add(time.Hour * -1)),
				},
				Status: v1.NodeStatus{
					NodeInfo: v1.NodeSystemInfo{
						KubeletVersion: "1.19.10",
					},
					Conditions: []v1.NodeCondition{
						{
							Type:   "Ready",
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			&v1.Node{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Node",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-b-1",
					Labels: map[string]string{
						"topology.kubernetes.io/zone": "us-east4-b",
					},
					CreationTimestamp: metav1.NewTime(time.Now().Add(time.Hour * -1)),
				},
				Status: v1.NodeStatus{
					NodeInfo: v1.NodeSystemInfo{
						KubeletVersion: "1.19.10",
					},
					Conditions: []v1.NodeCondition{
						{
							Type:   "Ready",
							Status: v1.ConditionTrue,
						},
					},
				},
			},
		)

		nodesCLI := nodes.NodesCLI{
			Client: clientset,
		}

		err := nodesCLI.Execute([]string{"node-a-1"}, writer)
		if err != nil {
			t.Errorf("Unexpected error: %v\n", err)
		}
		output, err := ioutil.ReadAll(writer)
		if err != nil {
			t.Errorf("Unexpected error reading output: %v\n", err)
		}

		expected := `NAME      STATUS  ROLES   AGE  VERSION  ZONE      
node-a-1  Ready   <none>  60m  1.19.10  us-east4-a
node-a-2  Ready   <none>  60m  1.19.10  us-east4-a
`

		if string(output) != expected {
			t.Errorf("Expected output to contain:\n%v\ngot:\n%v\n", expected, string(output))
		}
	})
}

func TestDefaultClient(t *testing.T) {
	t.Run("returns a configured Kubernetes client without error", func(t *testing.T) {
		workingDirectory, err := os.Getwd()
		if err != nil {
			t.Errorf("unexpected error getting working directory: %v", err)
		}
		_, err = nodes.DefaultClient(path.Join(workingDirectory, "../..", "test/test-kube-config"))
		if err != nil {
			t.Errorf("unexpected error getting default client: %v", err)
		}
	})
}
