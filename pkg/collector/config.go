package collector

import (
	"context"
	"sync"

	"github.com/Netcracker/network-latency-exporter/pkg/metrics"
	"github.com/Netcracker/network-latency-exporter/pkg/model"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
)

type ExporterConfig struct {
	LatencyTypes []string
	Namespace    string
	Mutex        sync.RWMutex
}

type Container struct {
	*ExporterConfig
	Exporter         *Exporter
	CollectorConfigs map[string]interface{}
	once             sync.Once
	logger           log.Logger
}

func NewConfigContainer(latencyTypes []string, namespace string, logger log.Logger) *Container {
	configHolder := &Container{
		ExporterConfig: &ExporterConfig{
			LatencyTypes: latencyTypes,
			Namespace:    namespace,
		},
		CollectorConfigs: make(map[string]interface{}),
		logger:           logger,
	}
	return configHolder
}

func (c *Container) UpdateTargets(ctx context.Context, targets metrics.PingHostList) {
	for _, latency := range c.ExporterConfig.LatencyTypes {
		switch latency {
		case string(NodeType):
			nConfig := c.CollectorConfigs[latency]
			nc := nConfig.(model.NodeCollector)
			nc.Targets = targets
			c.CollectorConfigs[latency] = nc
		case string(PodType):
			pConfig := c.CollectorConfigs[latency]
			pc := pConfig.(model.PodCollector)
			pc.Targets = targets
			c.CollectorConfigs[latency] = pc
		default:
			return
		}
	}
	_ = level.Info(c.Exporter.logger).Log("msg", "Updated targets")
}

func (c *Container) Initialize(ctx context.Context, packetsSent string, packetSize string, probeTimeout string, checkTargets []*metrics.CheckTarget, targets metrics.PingHostList, metricsPath string) (err error) {
	c.once.Do(func() {
		err = c.SetConfig(ctx, packetsSent, packetSize, probeTimeout, checkTargets, targets, metricsPath)
	})
	return
}

func (c *Container) SetConfig(ctx context.Context, packetsSent string, packetSize string, probeTimeout string, checkTargets []*metrics.CheckTarget, targets metrics.PingHostList, metricsPath string) error {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	if len(c.CollectorConfigs) != 0 {
		c.CollectorConfigs = make(map[string]interface{})
	}

	for _, latency := range c.ExporterConfig.LatencyTypes {
		switch latency {
		case string(NodeType):
			var nodeConfig model.NodeCollector
			nodeConfig.PacketsSent = packetsSent
			nodeConfig.PacketSize = packetSize
			nodeConfig.ProbeTimeout = probeTimeout
			nodeConfig.CheckTargets = checkTargets
			nodeConfig.Targets = targets
			nodeConfig.MetricsPath = metricsPath
			c.CollectorConfigs[latency] = nodeConfig
		case string(PodType):
			var podConfig model.PodCollector
			podConfig.PacketsSent = packetsSent
			podConfig.PacketSize = packetSize
			podConfig.ProbeTimeout = probeTimeout
			podConfig.CheckTargets = checkTargets
			podConfig.Targets = targets
			podConfig.MetricsPath = metricsPath
			c.CollectorConfigs[latency] = podConfig
		default:
			return errors.Errorf("Unknown collector type: %s", latency)
		}
	}
	return nil
}

func (c *Container) GetConfig(ctx context.Context, configType Type) interface{} {
	if cfg, found := c.CollectorConfigs[configType.String()]; found {
		return cfg
	}
	return nil
}
