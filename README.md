# kubectl nearby

A kubectl plugin that lists:

* pods on the same node as a given pod
* nodes in the same zone as a given node

## Installation

1. Download the latest version from the [Releases](https://github.com/leejones/kubectl-nearby/releases) page.
2. Place the `kubectl-nearby` binary in a directory that's included in your `$PATH`. For example: `/usr/local/bin`.

## Usage

### Nearby Pods

To list pods on the same node as a given pod:

```
kubectl nearby pods POD_NAME [OPTIONS]
```

By default, the output only shows pods from the same namespace as the given pod.

Options:

* `--namespace NAMESPACE` - The namespace for the given pod.
* `--all-namespaces` - The output will include pods from all namespaces on the same node as the given pod.
* `--kubeconfig` - The location of the kubeconfig file if it's not in a standard location.

### Nearby Nodes

To list nodes in the same zone as a given node:

```
kubectl nearby nodes NODE_NAME [OPTIONS]
```

kubectl-nearby uses the `topology.kubernetes.io/zone` label value to determine a node's zone.

Options:

* `--kubeconfig` - The location of the kubeconfig file if it's not in a standard location.

## Creating a new release

This is the standard release process for maintainers.

1. Install `gh` (the [GitHub CLI](https://cli.github.com)) if you don't have it already.
1. Note the most recent release version:

    ```bash
    gh release list --limit 1
    ```

1. Locally, checkout the `main` branch to the latest revision.
1. Set the version variable for the new release (in the form `vX.Y.Z`):

    ```bash
    read -p "Enter the new version (e.g., v1.2.3): " VERSION
    ```

1. Run the release script:

    ```bash
    bin/release $VERSION
    ```

1. Create the release using the GitHub CLI:

    ```bash
    gh release create $VERSION releases/${VERSION}/targets/*/*.tar.gz \
      --title "Release $VERSION" \
      --generate-notes
    ```
