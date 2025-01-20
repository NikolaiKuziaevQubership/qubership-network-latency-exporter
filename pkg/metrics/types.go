package metrics

import "strconv"

const (
	MeasurementName   = "network_latency"
	StatusOk          = 0
	StatusUnreachable = 1
)

type CheckTarget struct {
	Protocol string
	Port     string
	MtrKey   string
}

type MtrOutput struct {
	Report MtrOutputReport `json:"report"`
}

type MtrOutputReport struct {
	Mtr  MtrOutputMtr   `json:"mtr"`
	Hops []MtrOutputHop `json:"hubs"`
}

// MtrOutputMtr describes mtr run meta information
type MtrOutputMtr struct {
	// Source is a DNS name of source host
	Source string `json:"src"`
	// Destination is a IP address of destination host
	Destination string `json:"dst"`
	// PacketSize is a packet size in bytes
	PacketSize string `json:"psize"`
	// PacketsSent is an amount of packets which was sent
	PacketsSent int `json:"tests"`
}

// MtrOutputHop describes hop which packet uses during sending
type MtrOutputHop struct {
	// Number is a number of hop
	Number int `json:"count"`
	// Host is a name or IP address of hop
	Host string `json:"host"`
	// Sent is a total number a package sent
	Sent int `json:"Snt"`
	// Loss is a percent of lost packets
	Loss float64 `json:"Loss%"`
	// RttMean is an average mean of packets RTT
	RttMean float64 `json:"Avg"`
	// RttMin is an minimal packet RTT
	RttMin float64 `json:"Best"`
	// RttMax is a maximal packet RTT
	RttMax float64 `json:"Wrst"`
	// RttDeviation is a standard deviation of packets mean RTT
	RttDeviation float64 `json:"StDev"`
}

// NetworkLatencyMetric stores RTT and TTL data collected with `mtr` utility.
// The metric contains fields and tags which makes it applicable for InfluxDB.
type NetworkLatencyMetric struct {
	Tags   NetworkLatencyMetricTags
	Fields NetworkLatencyMetricFields
}

// NetworkLatencyMetricTags stores metric meta information.
type NetworkLatencyMetricTags struct {
	// Destination host name
	Dest string
	// Destination host IP address
	DestIp string
	// Protocol used for check
	Protocol string
	// Port used for check
	Port string
}

// NetworkLatencyMetricFields stores metric data.
type NetworkLatencyMetricFields struct {
	// Status - OK (mean host reached, but some packets can be lost) = 0 or UNREACHABLE = 1
	Status int
	// Total amount of UDP packets sent during probe
	TotalSent int
	// Total amount of UDP packets received during probe
	TotalReceived int
	// Mean round-trip time in milliseconds
	RttMean float64
	// Minimal round-trip time in milliseconds
	RttMin float64
	// Maximal round-trip time in milliseconds
	RttMax float64
	// Average round-trip time in milliseconds
	RttAvg float64
	// RttMean deviations of round-trip time in milliseconds
	RttDeviation float64
	// Number of hops in packet path
	HopsNum int
}

func NewNetworkLatencyMetric(dest string, destIp string, protocol string, port string, sent string) *NetworkLatencyMetric {
	m := &NetworkLatencyMetric{}
	m.Tags.Dest = dest
	m.Tags.DestIp = destIp
	m.Tags.Protocol = protocol
	m.Tags.Port = port
	m.Fields.TotalReceived = 0
	m.Fields.RttMean = 0.0
	m.Fields.RttMax = 0.0
	m.Fields.RttMin = 0.0
	m.Fields.RttDeviation = 0.0
	sentInt, _ := strconv.Atoi(sent)
	m.Fields.TotalSent = sentInt
	m.Fields.Status = StatusUnreachable
	return m
}

type PingHost struct {
	IPAddress string `yaml:"ipAddress"`
	Name      string `yaml:"name"`
}

// PingHostList stores list of ping targets to collect network latency metrics.
type PingHostList struct {
	Targets []PingHost `yaml:"targets"`
}
