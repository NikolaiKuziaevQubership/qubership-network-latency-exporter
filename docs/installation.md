# Installation guide

This section describes the `network-latency-exporter` installation process and requirements.

* [Installation guide](#installation-guide)
  * [Before Installation](#before-installation)
  * [Installation](#installation)
    * [Using Helm](#using-helm)
  * [Hardware Requirements](#hardware-requirements)
  * [Installation in Kubernetes / OpenShift cluster](#installation-in-kubernetes--openshift-cluster)
    * [RBAC](#rbac)
      * [ClusterRole](#clusterrole)
      * [ServiceAccount](#serviceaccount)
      * [ClusterRoleBinding](#clusterrolebinding)
      * [PodSecurityPolicy / SecurityContextConstraints](#podsecuritypolicy--securitycontextconstraints)
    * [Ports](#ports)
  * [Installation parameters](#installation-parameters)

## Before Installation

* Check [hardware requirements](#hardware-requirements)
* Check [roles requirements](#rbac)

## Installation

### Using Helm

To install with Helm, you must have [Helm v3](https://helm.sh/docs/intro/install/) installed.

See [Installation parameters](#installation-parameters) to customize your charts with `values` file.

Deploy with command

```bash
helm install network-latency-exporter ./charts/network-latency-exporter [-f <your_values_file>] -n <your_namespace>
```

Upgrade existing release with command

```bash
helm upgrade network-latency-exporter ./charts/network-latency-exporter [-f <your_values_file>] -n <your_namespace>
```

Uninstall `network-latency-exporter` with command

```bash
helm uninstall network-latency-exporter -n <your_namespace>
```

## Hardware Requirements

The hardware requirements depends on the service installation scheme.

## Installation in Kubernetes / OpenShift cluster

The service is installed as a `DaemonSet` so there are pods deployed on each `compute` node in cluster.

The RAM and CPU requirements for each pod:

| Resource | Requirement                                                                 |
| -------- | --------------------------------------------------------------------------- |
| RAM      | 10Mi + (5Mi \* (<number_of_nodes_in_cluster> - 1) \* <number_of_protocols>) |
| CPU      | 100m                                                                        |

For example, if there are 11 nodes in cluster the RAM request will be calculated as:

```ini
MEMORY_REQUEST = 10Mi + (5Mi * (11 - 1) * 3) = 160Mi
```

### RBAC

The `network-latency-exporter` requires a ClusterRole with ability to list cluster nodes and use a PodSecurityPolicy  
(in Kubernetes) or a SecurityContextConstraints (in OpenShift). Also, it requires to bind a ServiceAccount
to the ClusterRole with ClusterRoleBinding.

The necessary resources can be created during the service deploy if a user with cluster admin role is using for deploy.
Otherwise, an engineer must manually create the resources listed below.

#### ClusterRole

The ClusterRole should have a name which is equals to the service name (by default `network-latency-exporter`)
and access to list and watch cluster nodes, use PodSecurityPolicy and SecurityContextConstraints (both resources
can be placed in the ClusterRole regardless of the cluster type):

```yaml
apiVersion: v1
kind: ClusterRole
metadata:
  name: network-latency-exporter
rules:
  - apiGroups:
      - "security.openshift.io"
    resources:
      - securitycontextconstraints
    resourceNames:
      - network-latency-exporter
    verbs:
      - 'use'
  - apiGroups:
      - "extensions"
    resources:
      - podsecuritypolicies
    resourceNames:
      - network-latency-exporter
    verbs:
      - 'use'
  - apiGroups:
      - ""
    resources:
      - nodes
    verbs:
      - 'list'
      - 'watch'
```

#### ServiceAccount

The ServiceAccount should have a name which is equals to the service name (by default `network-latency-exporter`)
and should be bound to the created ClusterRole:

```yaml
kind: ServiceAccount
apiVersion: v1
metadata:
  name: network-latency-exporter
  labels:
    app.qubership.org/name: network-latency-exporter
```

#### ClusterRoleBinding

The ClusterRoleBinding should have at least one subject ServiceAccount with a name
(by default `network-latency-exporter`) and namespace (often for monitoring use namespace `monitoring`).
In the roleRef section it should have ClusterRole with a name (by default `network-latency-exporter`).
The ClusterRoleBinding binds created ServiceAccount to the ClusterRole.:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: network-latency-exporter
subjects:
- kind: ServiceAccount
  name: network-latency-exporter
  namespace: monitoring
roleRef:
  kind: ClusterRole
  name: network-latency-exporter
  apiGroup: rbac.authorization.k8s.io
```

#### PodSecurityPolicy / SecurityContextConstraints

The `network-latency-exporter` requires PodSecurityPolicy (or SecurityContextConstraints) for correct work because
the `mtr` tool cannot be used without root rights. So PSP and SCC below make it possible to run pods with exporter
with `securityContext.runAsUser: "0"` (root-user) parameter.

**NOTE**: you can use PSP / SCC that called `anyuid` instead of resources below.
The `anyuid` usually installed in the most of Kubernetes / OpenShift clusters by default.
In this case you have to change resourceNames in the ClusterRole to name of PSP / SCC that you want to use.

PodSecurityPolicy (only for Kubernetes cluster):

```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: network-latency-exporter
spec:
  privileged: false
  hostPID: false
  hostIPC: false
  hostNetwork: false
  volumes:
    - 'configMap'
    - 'secret'
  fsGroup:
    rule: 'MustRunAs'
    ranges:
      - min: 1
        max: 65535
  readOnlyRootFilesystem: false
  runAsUser:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'MustRunAs'
    ranges:
      - min: 1
        max: 65535
  allowPrivilegeEscalation: false
  seLinux:
    rule: 'RunAsAny'
  allowedCapabilities: []
```

SecurityContextConstraints (only for OpenShift cluster):

```yaml
apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: network-latency-exporter
priority: 0
users: []
groups: []
readOnlyRootFilesystem: false
requiredDropCapabilities: []
defaultAddCapabilities: []
allowedCapabilities: []
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: MustRunAs
allowPrivilegedContainer: false
allowHostDirVolumePlugin: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowHostIPC: false
allowPrivilegeEscalation: false
volumes:
  - configMap
  - secret
```

### Ports

The `network-latency-exporter` checks port `1` by default and this port must be opened on each node.
Otherwise, you should specify ports in the deploy parameter `checkTarget`.

## Installation parameters

This section describes the `network-latency-exporter` parameters for [install with Helm](#using-helm).
Parameters should be set in your `values.yaml` used when running `helm install` and `helm upgrade` commands.

To see all configurable options with default values, visit the
chart's [values.yaml](../../charts/network-latency-exporter/values.yaml), or run these configuration commands:

```bash
helm show values charts/network-latency-exporter
```

<!-- markdownlint-disable line-length -->
| Parameter                       | Type    | Mandatory | Default value                                                                | Description                                                                                                                                                                                                  |
| ------------------------------- | ------- | --------- | ---------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `nameOverride`                  | string  | no        | `network-latency-exporter`                                                   | Provide a name in place of network-latency-exporter for labels.                                                                                                                                              |
| `fullnameOverride`              | string  | no        | `<namespace>-network-latency-exporter`                                       | Provide a name to substitute for the full names of resources.                                                                                                                                                |
| `rbac.createClusterRole`        | boolean | no        | true                                                                         | Allow creating [ClusterRole](#clusterrole). If set to false, ClusterRole must be created manually.                                                                                                           |
| `rbac.createClusterRoleBinding` | boolean | no        | true                                                                         | Allow creating [ClusterRoleBinding](#clusterrolebinding). If set to false, ClusterRoleBinding must be created manually.                                                                                      |
| `createGrafanaDashboards`       | boolean | no        | true                                                                         | Allow creating Grafana Dashboards `Network Latency Overview` and `Network Latency Details`.                                                                                                                  |
| `serviceAccount.create`         | boolean | no        | true                                                                         | Allow creating ServiceAccount. If set to false, [ServiceAccount](#serviceaccount) must be created manually.                                                                                                  |
| `serviceAccount.name`           | boolean | no        | `network-latency-exporter`                                                   | Provide a name in place of network-latency-exporter for ServiceAccount.                                                                                                                                      |
| `image`                         | string  | yes       | `product/prod.platform.system.network-latency-exporter:master_latest`        | A docker image to use for network-latency-exporter daemonset.                                                                                                                                                |
| `resources`                     | object  | no        | `{requests: {cpu: 100m, memory: 128Mi}, limits: {cpu: 200m, memory: 256Mi}}` | The resources describes the compute resource requests and limits for single Pods.                                                                                                                            |
| `securityContext`               | object  | no        | `{runAsUser: "0", fsGroup: 2000}`                                            | SecurityContext holds pod-level security attributes.                                                                                                                                                         |
| `tolerations`                   | object  | no        | `[]`                                                                         | Tolerations allow the pods to schedule onto nodes with matching taints.                                                                                                                                      |
| `nodeSelector`                  | object  | no        | `{}`                                                                         | Allow to define which Nodes the Pods are scheduled on.                                                                                                                                                       |
| `affinity`                      | object  | no        | `{}`                                                                         | Pod's scheduling constraints.                                                                                                                                                                                |
| `discoverEnable`                | boolean | no        | true                                                                         | Allow enabling/disabling script for discovering nodes IP.                                                                                                                                                    |
| `requestTimeout`                | integer | no        | `3`                                                                          | Allow enabling/disabling script for discovering nodes IP.                                                                                                                                                    |
| `packetsNum`                    | integer | no        | `10`                                                                         | The number of packets to send per probe.                                                                                                                                                                     |
| `packetSize`                    | integer | no        | `64`                                                                         | The size of packet to sent in bytes.                                                                                                                                                                         |
| `checkTarget`                   | string  | no        | `"UDP:80,TCP:80,ICMP"`                                                       | The comma-separated list of network protocols and ports (separated by ':') via which packets will be sent. Supported protocols: UDP, TCP, ICMP. If no port is specified for protocol, port `1` will be used. |
| `timeout`                       | string  | no        | `100s`                                                                       | The metrics collection timeout. Can be calculated as, `TIMEOUT = 10s + (REQUEST_TIMEOUT * PACKETS_NUM * <NUMBER_OF_PROTOCOLS>)`.                                                                             |
| `serviceMonitor.enabled`        | boolean | no        | true                                                                         | If true, a ServiceMonitor is created for a `prometheus-operator`.                                                                                                                                            |
| `serviceMonitor.interval`       | string  | no        | `30s`                                                                        | Scraping interval for Prometheus.                                                                                                                                                                            |
| `additionalLabels`              | object  | no        | `[]`                                                                         | Allows specifying custom labels for DaemonSet of network-latency-exporter.                                                                                                                                   |
<!-- markdownlint-enable line-length -->
