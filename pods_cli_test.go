package main

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

func TestNewPodsCLINoArgs(t *testing.T) {
	args := []string{}
	want := noArgsError{}
	_, got := newPodsCLI(args)
	if got.Error() != want.Error() {
		t.Errorf("Expected newPodCLI with no args to return a errorNoArgs, but got: %v", got)
	}
}

func TestNewPodsCLIPodNameWithDefaults(t *testing.T) {
	args := []string{
		"nginx-abc123",
	}
	podsCLI, err := newPodsCLI(args)
	if err != nil {
		t.Errorf("Error creating new podsCLI: %v", err)
	}
	wantPodName := "nginx-abc123"
	gotpodName := podsCLI.podName
	if wantPodName != gotpodName {
		t.Errorf("podsCLI.podName should return %v, got: %v", wantPodName, gotpodName)
	}

	wantKubeconfig := ""
	gotKubeconfig := podsCLI.kubeconfig
	if wantKubeconfig != gotKubeconfig {
		t.Errorf("kubeconfig should default to: %v (an empty string), got: %v", wantKubeconfig, gotKubeconfig)
	}

	wantNamespace := "default"
	gotNamespace := podsCLI.namespace
	if wantNamespace != gotNamespace {
		t.Errorf("namespace should default to: %v, got: %v", wantNamespace, gotNamespace)
	}

	wantAllNamespaces := false
	gotAllNamespaces := podsCLI.allNamespaces
	if wantAllNamespaces != gotAllNamespaces {
		t.Errorf("podsCLI.namespace should return %v, got: %v", wantAllNamespaces, gotAllNamespaces)
	}
}

func TestNewPodsCLIPodCustomKubeconfig(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Errorf("Could not get working directory: %v", err)
	}
	args := []string{
		"nginx-abc123",
		"--kubeconfig",
		path.Join(workingDirectory, "test/test-kube-config"),
	}
	podsCLI, err := newPodsCLI(args)
	if err != nil {
		t.Errorf("Error creating new podsCLI: %v", err)
	}

	want := path.Join(workingDirectory, "test/test-kube-config")
	got := podsCLI.kubeconfig
	if want != got {
		t.Errorf("podsCLI.kubeconfig should return %v, got: %v", want, got)
	}
}

func TestNewPodsCLIPodAllNamespaces(t *testing.T) {
	args := []string{
		"nginx-abc123",
		"--all-namespaces",
	}
	podsCLI, err := newPodsCLI(args)
	if err != nil {
		t.Errorf("Error creating new podsCLI: %v", err)
	}

	want := true
	got := podsCLI.allNamespaces
	if want != got {
		t.Errorf("podsCLI.allNamespaces should return %v, got: %v", want, got)
	}
}

func TestNewPodsCLIPodCustomNamespace(t *testing.T) {
	args := []string{
		"nginx-abc123",
		"--namespace",
		"my-namespace",
	}
	podsCLI, err := newPodsCLI(args)
	if err != nil {
		t.Errorf("Error creating new podsCLI: %v", err)
	}

	want := "my-namespace"
	got := podsCLI.namespace
	if want != got {
		t.Errorf("podsCLI.namespace should return %v, got: %v", want, got)
	}
}

func TestNewPodsCLIHelp(t *testing.T) {
	argSets := [][]string{
		{"-h"},
		{"--help"},
	}

	for _, args := range argSets {
		want := helpRequestedError{}
		_, got := newPodsCLI(args)
		if want.Error() != got.Error() {
			t.Errorf("Expected newPodCLI with %v to return a helpRequestedError, but got: %v", args, got)
		}
	}
}

func TestNewPodsCLIInvalidFlag(t *testing.T) {
	args := []string{
		"--not-a-valid-flag",
	}
	want := "flag provided but not defined: -not-a-valid-flag"
	_, got := newPodsCLI(args)

	if want != got.Error() {
		t.Errorf("Expected newPodCLI with --not-a-valid-flag to return: %v, but got: %v", want, got)
	}
}

// TODO test podsCLI.clientset?

func TestColumnOuput(t *testing.T) {
	want := strings.Trim(`
NAMESPACE   NAME               READY
default     foo-bar-abc123     1/3  
production  baz-bat-db-def456  2/2  
`, "\n")
	input := [][]string{
		{"NAMESPACE", "NAME", "READY"},
		{"default", "foo-bar-abc123", "1/3"},
		{"production", "baz-bat-db-def456", "2/2"},
	}
	got, err := columnOutput(input)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if want != got {
		t.Errorf("Expected columnOutput to return:\n%v\n--- but got: ---\n%v", want, got)
	}
}

// ageOutput(d time.Duration) string
func TestAgeOutput(t *testing.T) {
	var testCases = []struct {
		input  string
		output string
	}{
		{"5s", "5s"},
		{"119s", "119s"},
		// >= 120s then use minutes and seconds
		{"121s", "2m1s"},
		{"9m59s", "9m59s"},
		// >= 10m, use minutes
		{"10m", "10m"},
		{"119m59s", "119m"},
		// >= 120m, use hours
		{"120m", "2h"},
		{"23h59m59s", "23h"},
		// >= 1d, use days
		{"24h0m1s", "1d"},
	}
	for _, testCase := range testCases {
		input, err := time.ParseDuration(testCase.input)
		if err != nil {
			t.Errorf("Unexpected error parsing testCase intput: %v, error: %v", testCase.input, err)
		}
		want := testCase.output
		got := ageOutput(input)
		if want != got {
			t.Errorf("Expected ageOutput(%v) to return: %v, but got: %v", input, want, got)
		}
	}
}
