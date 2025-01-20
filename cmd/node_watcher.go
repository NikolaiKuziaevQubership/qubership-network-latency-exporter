package main

import (
	"context"
	"fmt"

	"github.com/Netcracker/network-latency-exporter/pkg/collector"
	"github.com/Netcracker/network-latency-exporter/pkg/utils"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type nodeWatcher struct {
	ctx    context.Context
	logger log.Logger
}

func (fw *nodeWatcher) watch(logger log.Logger, watcher watch.Interface, cfgCont *collector.Container, ctx context.Context) {
	for event := range watcher.ResultChan() {
		_ = level.Info(logger).Log("msg", fmt.Sprintf("Event occurred: %v on node %v", event.Type, event.Object.(*v1.Node).Name))
		if event.Type == watch.Added || event.Type == watch.Modified || event.Type == watch.Deleted {
			targets := collector.Discover(logger)
			if targets != nil {
				targets = utils.ValidateTargets(logger, targets)
				cfgCont.UpdateTargets(ctx, *targets)
			}
		}
	}
}
