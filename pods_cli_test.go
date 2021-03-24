package main

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"k8s.io/client-go/util/homedir"
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

	home := homedir.HomeDir()
	wantKubeconfig := filepath.Join(home, ".kube", "config")
	gotKubeconfig := podsCLI.kubeconfig
	if wantKubeconfig != gotKubeconfig {
		t.Errorf("kubeconfig should default to: %v, got: %v", wantKubeconfig, gotKubeconfig)
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
