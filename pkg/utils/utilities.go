package utils

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/Netcracker/network-latency-exporter/pkg/metrics"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetEnvWithDefaultValue(key string, defaultValue string) string {
	value, found := os.LookupEnv(key)
	if !found {
		return defaultValue
	}
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func GetClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func ValidateTargets(logger log.Logger, targets *metrics.PingHostList) *metrics.PingHostList {
	res := &metrics.PingHostList{}
	for _, t := range targets.Targets {
		if t.IPAddress != "" && net.ParseIP(t.IPAddress) != nil {
			res.Targets = append(res.Targets, t)
		} else {
			_ = level.Warn(logger).Log(fmt.Sprintf("Skip the invalid ping target: {ipAddress: %s, name: %s}. "+
				"The `ipAddress` should be a non-empty valid IP address", t.IPAddress, t.Name))
		}
	}
	return res
}

func GetNamespace() string {
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return "monitoring"
}

// add HSTS header
func AddHSTSHeader(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Scheme == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		handler.ServeHTTP(w, r)
	})
}
