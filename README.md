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

1. Install the [GitHub CLI](https://cli.github.com) if you don't have it already.
1. Note the most recent release version:

    ```bash
    gh release list --limit 1
    ```

1. Locally, checkout the `main` branch to the latest revision.
1. Run `bin/release VERSION` where version is in the form `vX.Y.Z` and is the next logical semantic version based on the changes.
1. Create the release using the GitHub CLI:

   ```bash
   gh release create vX.Y.Z releases/vX.Y.Z/targets/*/*.tar.gz \
     --title "Release vX.Y.Z" \
     --generate-notes
   ```
