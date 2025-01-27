package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/Netcracker/network-latency-exporter/pkg/collector"
	"github.com/Netcracker/network-latency-exporter/pkg/metrics"
	"github.com/Netcracker/network-latency-exporter/pkg/utils"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	versionCollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func init() {
	prometheus.MustRegister(versionCollector.NewCollector("network_latency_exporter"))
}

func main() {
	var (
		webConfig    = webflag.AddFlags(kingpin.CommandLine, ":9273")
		packetsSent  = utils.GetEnvWithDefaultValue("PACKETS_NUM", "10")
		packetSize   = utils.GetEnvWithDefaultValue("PACKET_SIZE", "1500")
		protocolsStr = utils.GetEnvWithDefaultValue("CHECK_TARGET", "ICMP")
		probeTimeout = utils.GetEnvWithDefaultValue("REQUEST_TIMEOUT", "3")
		latencyTypes = utils.GetEnvWithDefaultValue("LATENCY_TYPES", "node_collector")
		metricsPath  = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		maxRequests = kingpin.Flag(
			"web.max-requests",
			"Maximum number of parallel scrape requests. Use 0 to disable.",
		).Default("40").Int()
	)

	promLogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promLogConfig)
	kingpin.Version(version.Print("network_latency_exporter"))
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promLogConfig)

	_ = os.Setenv("LOG_LEVEL", promLogConfig.Level.String())

	_ = level.Info(logger).Log("msg", fmt.Sprintf("Starting network_latency_exporter: %s", version.Info()))
	_ = level.Info(logger).Log("msg", fmt.Sprintf("Build context: %s", version.BuildContext()))

	namespace := utils.GetNamespace()
	_ = level.Info(logger).Log("msg", fmt.Sprintf("Namespace: %s", namespace))

	baseCtx, cancel := context.WithCancel(context.Background())
	ctx := context.WithValue(baseCtx, collector.ContextKey, "main")

	rCfg, _ := ctrl.GetConfig()
	var clientSet *kubernetes.Clientset
	if rCfg != nil {
		clientSet = kubernetes.NewForConfigOrDie(rCfg)
	}

	var checkTargets []*metrics.CheckTarget
	for _, p := range strings.Split(strings.TrimSpace(protocolsStr), ",") {
		protocolAndPort := strings.Split(p, ":")
		checkTarget := &metrics.CheckTarget{}
		if protocolAsFlag, ok := collector.ProtocolToMtrFlag[protocolAndPort[0]]; ok {
			checkTarget.Protocol = protocolAndPort[0]
			checkTarget.MtrKey = protocolAsFlag
		} else {
			_ = level.Warn(logger).Log("msg", fmt.Sprintf("Skip incorrect or unsupported protocol %s", p))
			continue
		}
		if len(protocolAndPort) == 2 {
			checkTarget.Port = protocolAndPort[1]
		} else {
			checkTarget.Port = "1"
		}
		checkTargets = append(checkTargets, checkTarget)
	}

	targets := collector.Discover(logger)
	if targets != nil {
		targets = utils.ValidateTargets(logger, targets)
		latencies := strings.Split(latencyTypes, ",")
		cfgCont := collector.NewConfigContainer(latencies, namespace, logger)
		if err := cfgCont.Initialize(ctx, packetsSent, packetSize, probeTimeout, checkTargets, *targets, *metricsPath); err != nil {
			_ = level.Error(logger).Log("msg", "Initialization failed", "err", err)
			os.Exit(1)
		}

		var enabledCollectors []collector.Collector
		for collectorName, enabled := range collector.GetCollectorStates() {
			if _, found := cfgCont.CollectorConfigs[string(collector.AsType(collectorName))]; found && enabled {
				_ = level.Info(logger).Log("msg", fmt.Sprintf("Collector enabled from main: %s", collectorName))
				c, err := collector.GetCollector(collectorName, logger)
				if err != nil {
					_ = level.Error(logger).Log("msg", fmt.Sprintf("Couldn't get collector: %s", collectorName), "err", err)
					continue
				}
				enabledCollectors = append(enabledCollectors, c)
			}
		}

		for _, coll := range enabledCollectors {
			if cfg := cfgCont.GetConfig(ctx, coll.Type()); cfg != nil {
				err := coll.Initialize(ctx, cfg)
				if err != nil {
					_ = level.Error(logger).Log("msg", fmt.Sprintf("Can't initialize collector: %s", coll.Name()), "err", err)
				}
			}
		}
		exporter := collector.New(ctx, collector.NewMetrics(), enabledCollectors, logger)
		cfgCont.Exporter = exporter

		watcher, err := clientSet.CoreV1().Nodes().Watch(context.TODO(), metav1.ListOptions{})
		if err != nil {
			_ = level.Error(logger).Log("msg", err.Error())
			syscall.Exit(1)
		}

		defer func(watcher watch.Interface) {
			watcher.Stop()
		}(watcher)

		nw := &nodeWatcher{
			ctx:    ctx,
			logger: logger,
		}

		go nw.watch(logger, watcher, cfgCont, ctx)

		// register exporter only once
		err = prometheus.Register(exporter)
		if err != nil {
			if !prometheus.Unregister(exporter) {
				_ = level.Error(logger).Log("msg", "Exporter can't be unregistered")
				return
			}
			prometheus.MustRegister(exporter)
		}

		metricHandlerFunc := collector.MetricHandler(exporter, *maxRequests, logger)
		http.Handle(*metricsPath, utils.AddHSTSHeader(promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, metricHandlerFunc)))
		http.Handle("/-/ready", utils.AddHSTSHeader(readinessChecker()))
		http.Handle("/-/healthy", utils.AddHSTSHeader(healthChecker()))

		srvBaseCtx := context.WithValue(context.Background(), collector.ContextKey, "http")
		srv := &http.Server{
			BaseContext: func(_ net.Listener) context.Context {
				return srvBaseCtx
			},
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
		}
		sd := &shutdown{
			srv:     srv,
			logger:  logger,
			ctx:     context.WithValue(context.Background(), collector.ContextKey, "shutdown"),
			timeout: 30 * time.Second,
		}
		go sd.listen()
		_ = level.Info(logger).Log("msg", fmt.Sprintf("Starting server on address %s", srv.Addr))
		exit := web.ListenAndServe(srv, webConfig, logger)

		cancel()
		if !errors.Is(exit, http.ErrServerClosed) {
			_ = level.Error(logger).Log("msg", "Failed to start application", "err", exit)
		}
		_ = level.Info(logger).Log("msg", "Server is shut down")
	} else {
		_ = level.Info(logger).Log("msg", "Discovery is disabled")
	}

}

func healthChecker() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			_, _ = w.Write([]byte("OK"))
		},
	)
}

func readinessChecker() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			_, _ = w.Write([]byte("OK"))
		})
}
