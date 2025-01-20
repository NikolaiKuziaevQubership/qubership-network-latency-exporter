# Metrics list

| Name                       | Type, Unit | Description                                                    |
| -------------------------- | ---------- | -------------------------------------------------------------- |
| network_latency_status     | gauge      | Status of network latency. 0 if successful, 1 if unsuccessful. |
| network_latency_sent       | gauge      | The total number of packets sent.                              |
| network_latency_received   | gauge      | The total number of packets received.                          |
| network_latency_rtt_min    | gauge      | Best round trip time (RTT)                                     |
| network_latency_rtt_max    | gauge      | Worst round trip time (RTT).                                   |
| network_latency_rtt_mean   | gauge      | Average mean of RTT packets.                                   |
| network_latency_rtt_stddev | gauge      | Standard deviation of packets mean RTT.                        |
| network_latency_hops_num   | gauge      | Number of hops in packet path.                                 |
