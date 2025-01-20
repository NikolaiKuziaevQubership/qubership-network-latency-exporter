package collector

import (
	"context"
	"fmt"

	"github.com/Netcracker/network-latency-exporter/pkg/metrics"
	"github.com/Netcracker/network-latency-exporter/pkg/utils"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getClusterNodes returns list of cluster nodes.
func getClusterNodes() ([]corev1.Node, error) {
	// Creates the in-cluster client
	clientSet, err := utils.GetClientset()
	if err != nil {
		return nil, err
	}

	// Reads nodes list
	nodes, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return nodes.Items, nil
}

func Discover(logger log.Logger) *metrics.PingHostList {
	if utils.GetEnvWithDefaultValue("DISCOVER_ENABLE", "true") == "true" {
		_ = level.Debug(logger).Log("msg", "Discovering cluster nodes as ping targets")
		rawNodes, err := getClusterNodes()
		if err != nil {
			_ = level.Debug(logger).Log("msg", fmt.Sprintf("Error getting cluster nodes: %v", err))
			return nil
		}

		// Parse nodes as targets
		targets := &metrics.PingHostList{}

		for _, n := range rawNodes {
			nodeAddress := ""
			nodeName := ""
			for _, a := range n.Status.Addresses {
				if a.Type == corev1.NodeInternalIP {
					nodeAddress = a.Address
				}
				if a.Type == corev1.NodeHostName {
					nodeName = a.Address
				}
			}

			if nodeAddress != "" {
				// Skip current node
				if nodeName != utils.GetEnvWithDefaultValue("NODE_NAME", "localhost") {
					_ = level.Debug(logger).Log("msg", fmt.Sprintf("Discovered node: {ipAddress: %s, name: %s}", nodeAddress, nodeName))
					targets.Targets = append(targets.Targets, metrics.PingHost{IPAddress: nodeAddress, Name: nodeName})
				}
			}
		}
		return targets
	} else {
		_ = level.Info(logger).Log("msg", "Skip discovering. Script disabled.")
		return nil
	}
}
