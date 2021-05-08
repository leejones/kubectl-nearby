# kubectl nearby

A kubectl plugin that lists pods on the same node as the given pod.

## Installation

1. Download the latest version from the [releases](/releases) page.
2. Place the `kubectl-nearby` binary in a directory that's included in your `$PATH`. For example: `/usr/local/bin`.

## Usage

To list pods on the same node as a given pod:

```
kubectl nearby pods POD_NAME [OPTIONS]
```

By default, the output only shows pods from the same namespace as the given pod.

Options:

* `--namespace NAMESPACE` - The namespace for the given pod.
* `--all-namespaces` - The output will include pods from all namespaces on the same node as the given pod.
