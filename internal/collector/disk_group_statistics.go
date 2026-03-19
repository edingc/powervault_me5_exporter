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

type diskGroupStatsDescs struct {
	avgReadRspTimeMicroseconds  *prometheus.Desc
	avgRspTimeMicroseconds      *prometheus.Desc
	avgWriteRspTimeMicroseconds *prometheus.Desc
	bytesPerSecond              *prometheus.Desc
	dataReadBytesTotal          *prometheus.Desc
	dataWrittenBytesTotal       *prometheus.Desc
	iops                        *prometheus.Desc
	pagesAllocatedPerMinute     *prometheus.Desc
	pagesDeallocatedPerMinute   *prometheus.Desc
	readsTotal                  *prometheus.Desc
	writesTotal                 *prometheus.Desc
}

func newDiskGroupStatsDescs(label func(string, string, string, ...string) *prometheus.Desc) diskGroupStatsDescs {
	commonLabels := []string{"name", "serial"}

	return diskGroupStatsDescs{
		avgReadRspTimeMicroseconds:  label("disk_group", "avg_read_rsp_time_microseconds", "Average response time in microseconds for all read operations, calculated over the interval since these statistics were last requested or reset.", commonLabels...),
		avgRspTimeMicroseconds:      label("disk_group", "avg_rsp_time_microseconds", "Average response time in microseconds for read and write operations, calculated over the interval since these statistics were last requested or reset.", commonLabels...),
		avgWriteRspTimeMicroseconds: label("disk_group", "avg_write_rsp_time_microseconds", "Average response time in microseconds for all write operations, calculated over the interval since these statistics were last requested or reset.", commonLabels...),
		bytesPerSecond:              label("disk_group", "bytes_per_second", "The data transfer rate, in bytes per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.", commonLabels...),
		dataReadBytesTotal:          label("disk_group", "data_read_bytes_total", "Amount of data read since these statistics were last reset or since the controller was restarted, in bytes.", commonLabels...),
		dataWrittenBytesTotal:       label("disk_group", "data_written_bytes_total", "Amount of data written since these statistics were last reset or since the controller was restarted, in bytes.", commonLabels...),
		iops:                        label("disk_group", "iops", "Input/output operations per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.", commonLabels...),
		pagesAllocatedPerMinute:     label("disk_group", "pages_allocated_per_minute", "The rate, in pages per minute, at which pages are allocated to volumes in the disk group because they need more space to store data.", commonLabels...),
		pagesDeallocatedPerMinute:   label("disk_group", "pages_deallocated_per_minute", "The rate, in pages per minute, at which pages are deallocated from volumes in the disk group because they no longer need the space to store data.", commonLabels...),
		readsTotal:                  label("disk_group", "reads_total", "Number of read operations since these statistics were last reset or since the controller was restarted.", commonLabels...),
		writesTotal:                 label("disk_group", "writes_total", "Number of write operations since these statistics were last reset or since the controller was restarted.", commonLabels...),
	}
}

func (d diskGroupStatsDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.avgReadRspTimeMicroseconds
	ch <- d.avgRspTimeMicroseconds
	ch <- d.avgWriteRspTimeMicroseconds
	ch <- d.bytesPerSecond
	ch <- d.dataReadBytesTotal
	ch <- d.dataWrittenBytesTotal
	ch <- d.iops
	ch <- d.pagesAllocatedPerMinute
	ch <- d.pagesDeallocatedPerMinute
	ch <- d.readsTotal
	ch <- d.writesTotal
}

func (c *ME5Collector) CollectDiskGroupStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		DiskGroupStatistics []struct {
			AvgReadRspTime           float64 `json:"avg-read-rsp-time"`
			AvgRspTime               float64 `json:"avg-rsp-time"`
			AvgWriteRspTime          float64 `json:"avg-write-rsp-time"`
			BytesPerSecond           float64 `json:"bytes-per-second-numeric"`
			DataReadNumeric          float64 `json:"data-read-numeric"`
			DataWrittenNumeric       float64 `json:"data-written-numeric"`
			IOPS                     float64 `json:"iops"`
			Name                     string  `json:"name"`
			NumberOfReads            float64 `json:"number-of-reads"`
			NumberOfWrites           float64 `json:"number-of-writes"`
			SerialNumber             string  `json:"serial-number"`
			DiskGroupStatisticsPaged []struct {
				PagesAllocPerMinute   float64 `json:"pages-alloc-per-minute"`
				PagesDeallocPerMinute float64 `json:"pages-dealloc-per-minute"`
			} `json:"disk-group-statistics-paged,omitempty"`
		} `json:"disk-group-statistics"`
	}

	if err := c.client.Get(ctx, "/show/disk-group-statistics", &resp); err != nil {
		slog.Error("failed to fetch disk group statistics from API", "endpoint", "/show/disk-group-statistics", "error", err)
		return err
	}

	for _, s := range resp.DiskGroupStatistics {
		name := s.Name
		sn := s.SerialNumber

		ch <- prometheus.MustNewConstMetric(c.diskGroupStats.avgReadRspTimeMicroseconds, prometheus.GaugeValue, s.AvgReadRspTime, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroupStats.avgRspTimeMicroseconds, prometheus.GaugeValue, s.AvgRspTime, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroupStats.avgWriteRspTimeMicroseconds, prometheus.GaugeValue, s.AvgWriteRspTime, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroupStats.bytesPerSecond, prometheus.GaugeValue, s.BytesPerSecond, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroupStats.dataReadBytesTotal, prometheus.CounterValue, s.DataReadNumeric, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroupStats.dataWrittenBytesTotal, prometheus.CounterValue, s.DataWrittenNumeric, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroupStats.iops, prometheus.GaugeValue, s.IOPS, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroupStats.readsTotal, prometheus.CounterValue, s.NumberOfReads, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroupStats.writesTotal, prometheus.CounterValue, s.NumberOfWrites, name, sn)

		for _, p := range s.DiskGroupStatisticsPaged {
			ch <- prometheus.MustNewConstMetric(c.diskGroupStats.pagesAllocatedPerMinute, prometheus.GaugeValue, p.PagesAllocPerMinute, name, sn)
			ch <- prometheus.MustNewConstMetric(c.diskGroupStats.pagesDeallocatedPerMinute, prometheus.GaugeValue, p.PagesDeallocPerMinute, name, sn)
		}
	}

	return nil
}
