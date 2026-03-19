// Copyright 2026 Cody Eding
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

type hostPortStatsDescs struct {
	avgRspTimeMicroseconds *prometheus.Desc
	bytesPerSecond         *prometheus.Desc
	dataReadBytesTotal     *prometheus.Desc
	dataWrittenBytesTotal  *prometheus.Desc
	iops                   *prometheus.Desc
	queueDepth             *prometheus.Desc
	readsTotal             *prometheus.Desc
	writesTotal            *prometheus.Desc
}

func newHostPortStatsDescs(label func(string, string, string, ...string) *prometheus.Desc) hostPortStatsDescs {
	return hostPortStatsDescs{
		avgRspTimeMicroseconds: label("host_port", "avg_rsp_time_microseconds", "Average response time in microseconds for read and write operations, calculated over the interval since these statistics were last requested or reset.", "port"),
		bytesPerSecond:         label("host_port", "bytes_per_second", "The data transfer rate, in bytes per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.", "port"),
		dataReadBytesTotal:     label("host_port", "data_read_bytes_total", "Amount of data read since these statistics were last reset or since the controller was restarted.", "port"),
		dataWrittenBytesTotal:  label("host_port", "data_written_bytes_total", "Amount of data written since these statistics were last reset or since the controller was restarted.", "port"),
		iops:                   label("host_port", "iops", "Input/output operations per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.", "port"),
		queueDepth:             label("host_port", "queue_depth", "The number of pending I/O operations currently being serviced.", "port"),
		readsTotal:             label("host_port", "reads_total", "Number of read operations since these statistics were last reset or since the controller was restarted.", "port"),
		writesTotal:            label("host_port", "writes_total", "Number of write operations since these statistics were last reset or since the controller was restarted.", "port"),
	}
}

func (d hostPortStatsDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.avgRspTimeMicroseconds
	ch <- d.bytesPerSecond
	ch <- d.dataReadBytesTotal
	ch <- d.dataWrittenBytesTotal
	ch <- d.iops
	ch <- d.queueDepth
	ch <- d.readsTotal
	ch <- d.writesTotal
}

func (c *ME5Collector) CollectHostPortStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		HostPortStatistics []struct {
			AvgRspTime         float64 `json:"avg-rsp-time"`
			BytesPerSecond     float64 `json:"bytes-per-second-numeric"`
			DataReadNumeric    float64 `json:"data-read-numeric"`
			DataWrittenNumeric float64 `json:"data-written-numeric"`
			DurableID          string  `json:"durable-id"`
			IOPS               float64 `json:"iops"`
			NumberOfReads      float64 `json:"number-of-reads"`
			NumberOfWrites     float64 `json:"number-of-writes"`
			QueueDepth         float64 `json:"queue-depth"`
		} `json:"host-port-statistics"`
	}

	if err := c.client.Get(ctx, "/show/host-port-statistics", &resp); err != nil {
		slog.Error("failed to fetch host-port-statistics from API", "endpoint", "/show/host-port-statistics", "error", err)
		return err
	}

	for _, s := range resp.HostPortStatistics {
		port := s.DurableID
		ch <- prometheus.MustNewConstMetric(c.hostPortStats.avgRspTimeMicroseconds, prometheus.GaugeValue, s.AvgRspTime, port)
		ch <- prometheus.MustNewConstMetric(c.hostPortStats.bytesPerSecond, prometheus.GaugeValue, s.BytesPerSecond, port)
		ch <- prometheus.MustNewConstMetric(c.hostPortStats.dataReadBytesTotal, prometheus.CounterValue, s.DataReadNumeric, port)
		ch <- prometheus.MustNewConstMetric(c.hostPortStats.dataWrittenBytesTotal, prometheus.CounterValue, s.DataWrittenNumeric, port)
		ch <- prometheus.MustNewConstMetric(c.hostPortStats.iops, prometheus.GaugeValue, s.IOPS, port)
		ch <- prometheus.MustNewConstMetric(c.hostPortStats.queueDepth, prometheus.GaugeValue, s.QueueDepth, port)
		ch <- prometheus.MustNewConstMetric(c.hostPortStats.readsTotal, prometheus.CounterValue, s.NumberOfReads, port)
		ch <- prometheus.MustNewConstMetric(c.hostPortStats.writesTotal, prometheus.CounterValue, s.NumberOfWrites, port)
	}

	return nil
}
