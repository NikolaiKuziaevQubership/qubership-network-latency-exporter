# Network Latency Exporter

* [Network Latency Exporter](#network-latency-exporter)
  * [Overview](#overview)
  * [Documents](#documents)
  * [Releases](#releases)
  * [How to start](#how-to-start)
    * [Build](#build)
      * [Local build](#local-build)
    * [Deploy to k8s](#deploy-to-k8s)
      * [Pure helm](#pure-helm)
    * [Smoke tests](#smoke-tests)
    * [How to debug](#how-to-debug)
    * [How to troubleshoot](#how-to-troubleshoot)
  * [Evergreen strategy](#evergreen-strategy)

## Overview

The `network-latency-exporter` is a service which collects RTT and TTL metrics
for the list of target hosts and sends collected data to an InfluxDB instance or Prometheus.

It is possible to use `UDP`, `TCP` or `ICMP` network protocols to sent package during probes.

The service collects metrics with `mtr` tool which accumulates functionality of `ping` and `traceroute` tools.
Target hosts can be discovered automatically by retrieving all k8s cluster nodes.

The list of metrics collected by `network-latency-exporter` can be found [here](/docs/public/metrics.md)

## Documents

* [Installation](docs/public/installation.md)

## Releases

Not applicable.

## How to start

### Build

#### Local build

1. Use WSL or Linux VM
2. Run commands:

    ```bash
    sh .\build.sh
    ```

### Deploy to k8s

#### Pure helm

1. Create and fill the file `charts/network-latency-exporter/custom.values.yaml`
2. Run command:

    ```bash
    helm install -n <namespace> network-latency-exporter charts/network-latency-exporter -f charts/network-latency-exporter/custom.values.yaml
    ```

More details see in the [Installation Guide: Helm](docs/public/installation.md#using-helm).

### Smoke tests

There are no smoke tests.

### How to debug

It's possible to build and run local build of network-latency-exporter with connection to the Kubernetes.
It can be useful in case if you want to debug some functionality of network-latency-exporter.

To do it need:

1. Install `kubectl`, see [https://kubernetes.io/docs/tasks/tools/#kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
2. Configure `kubectl` context to connect to necessary Kubernetes.

    General template that allow you configure a lot of contexts and switch between them:

    ```bash
    kubectl config set-cluster <namespace>/<url>/<user> --server=https://<kube_api_url>:<kube_api_port> --insecure-skip-tls-verify=true
    kubectl config set-context <namespace>/<url>/<user> --cluster=<namespace>/<url>/<user>
    kubectl config set-context <namespace>/<url>/<user> --user=<namespace>/<url>/<user>
    kubectl config set-credentials <namespace>/<url>/<user> --token=<token>
    kubectl config use-context <namespace>/<url>/<user>
    ```

    In future context can be switched using the command:

    ```bash
    kubectl config use-context <namespace>/<url>/<user>
    ```

    Or using the special tools like `kubectx` and `kubens` [https://github.com/ahmetb/kubectx](https://github.com/ahmetb/kubectx)

3. Configure your IDE (VSCode) to build and run network-latency-exporter in debug mode:

    ```json
    {
        "version": "0.2.0",
        "configurations": [
            {
                "name": "Launch Package",
                "type": "go",
                "request": "launch",
                "mode": "auto",
                "program": "${workspaceFolder}/cmd/main.go"
            }
        ]
    }
    ```

4. Scale down to 0 replicas or remove network-latency-exporter in the Kubernetes to avoid conflicts
5. Run the local network-latency-exporter in debug mode

### How to troubleshoot

Check details in [Troubleshooting Guide](docs/troubleshooting.md).

## Evergreen strategy

To keep the component up to date, the following activities should be performed regularly:

* vulnerabilities fixing
* bug-fixing, improvement and feature implementation for exporter
