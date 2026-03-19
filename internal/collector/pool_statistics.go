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

type poolStatsDescs struct {
	avgReadRspTimeMicroseconds  *prometheus.Desc
	avgRspTimeMicroseconds      *prometheus.Desc
	avgWriteRspTimeMicroseconds *prometheus.Desc
	bytesPerSecond              *prometheus.Desc
	dataReadBytesTotal          *prometheus.Desc
	dataWrittenBytesTotal       *prometheus.Desc
	iops                        *prometheus.Desc
	numColdPageMoves            *prometheus.Desc
	numHotPageMoves             *prometheus.Desc
	pagesAllocPerMinute         *prometheus.Desc
	pagesDeallocPerMinute       *prometheus.Desc
	pagesUnmapPerMinute         *prometheus.Desc
	readsTotal                  *prometheus.Desc
	writesTotal                 *prometheus.Desc
}

func newPoolStatsDescs(label func(string, string, string, ...string) *prometheus.Desc) poolStatsDescs {
	return poolStatsDescs{
		avgReadRspTimeMicroseconds:  label("pool", "avg_read_rsp_time_microseconds", "The average response time, in microseconds, for read operations since the last sampling time.", "pool", "serial"),
		avgRspTimeMicroseconds:      label("pool", "avg_rsp_time_microseconds", "The average response time, in microseconds, for read and write operations since the last sampling time.", "pool", "serial"),
		avgWriteRspTimeMicroseconds: label("pool", "avg_write_rsp_time_microseconds", "The average response time, in microseconds, for write operations since the last sampling time.", "pool", "serial"),
		bytesPerSecond:              label("pool", "bytes_per_second", "The data transfer rate, in bytes per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.", "pool", "serial"),
		dataReadBytesTotal:          label("pool", "data_read_bytes_total", "The amount of data read since these statistics were last reset or since the controller was restarted.", "pool", "serial"),
		dataWrittenBytesTotal:       label("pool", "data_written_bytes_total", "The amount of data written since these statistics were last reset or since the controller was restarted.", "pool", "serial"),
		iops:                        label("pool", "iops", "The number of input/output operations per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.", "pool", "serial"),
		numColdPageMoves:            label("pool", "num_cold_page_moves", "The number of 'cold' pages promoted from lower tiers to higher tiers since statistics were last reset.", "pool", "serial"),
		numHotPageMoves:             label("pool", "num_hot_page_moves", "The number of 'hot' pages promoted from lower tiers to higher tiers since statistics were last reset.", "pool", "serial"),
		pagesAllocPerMinute:         label("pool", "pages_alloc_per_minute", "The rate, in pages per minute, at which pages are allocated to volumes in the pool because they need more space to store data.", "pool", "serial"),
		pagesDeallocPerMinute:       label("pool", "pages_dealloc_per_minute", "The rate, in pages per minute, at which pages are deallocated from volumes in the pool because they no longer need the space to store data.", "pool", "serial"),
		pagesUnmapPerMinute:         label("pool", "pages_unmap_per_minute", "The number of 4 MB pages that host systems have unmapped per minute, through use of the SCSI UNMAP command, to free storage space as a result of deleting files or formatting volumes on the host.", "pool", "serial"),
		readsTotal:                  label("pool", "reads_total", "The number of read operations since these statistics were last reset or since the controller was restarted.", "pool", "serial"),
		writesTotal:                 label("pool", "writes_total", "The number of write operations since these statistics were last reset or since the controller was restarted.", "pool", "serial"),
	}
}

func (d poolStatsDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.avgReadRspTimeMicroseconds
	ch <- d.avgRspTimeMicroseconds
	ch <- d.avgWriteRspTimeMicroseconds
	ch <- d.bytesPerSecond
	ch <- d.dataReadBytesTotal
	ch <- d.dataWrittenBytesTotal
	ch <- d.iops
	ch <- d.numColdPageMoves
	ch <- d.numHotPageMoves
	ch <- d.pagesAllocPerMinute
	ch <- d.pagesDeallocPerMinute
	ch <- d.pagesUnmapPerMinute
	ch <- d.readsTotal
	ch <- d.writesTotal
}

func (c *ME5Collector) CollectPoolStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		PoolStatistics []struct {
			NumColdPageMoves      float64 `json:"num-cold-page-moves"`
			NumHotPageMoves       float64 `json:"num-hot-page-moves"`
			PagesAllocPerMinute   float64 `json:"pages-alloc-per-minute"`
			PagesDeallocPerMinute float64 `json:"pages-dealloc-per-minute"`
			PagesUnmapPerMinute   float64 `json:"pages-unmap-per-minute"`
			Pool                  string  `json:"pool"`
			ResettableStatistics  []struct {
				AvgReadRspTime     float64 `json:"avg-read-rsp-time"`
				AvgRspTime         float64 `json:"avg-rsp-time"`
				AvgWriteRspTime    float64 `json:"avg-write-rsp-time"`
				BytesPerSecond     float64 `json:"bytes-per-second-numeric"`
				DataReadNumeric    float64 `json:"data-read-numeric"`
				DataWrittenNumeric float64 `json:"data-written-numeric"`
				IOPS               float64 `json:"iops"`
				NumberOfReads      float64 `json:"number-of-reads"`
				NumberOfWrites     float64 `json:"number-of-writes"`
			} `json:"resettable-statistics"`
			SerialNumber string `json:"serial-number"`
		} `json:"pool-statistics"`
	}

	if err := c.client.Get(ctx, "/show/pool-statistics", &resp); err != nil {
		slog.Error("failed to fetch pool-statistics from API", "endpoint", "/show/pool-statistics", "error", err)
		return err
	}

	for _, s := range resp.PoolStatistics {
		ch <- prometheus.MustNewConstMetric(c.poolStats.numColdPageMoves, prometheus.CounterValue, s.NumColdPageMoves, s.Pool, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.poolStats.numHotPageMoves, prometheus.CounterValue, s.NumHotPageMoves, s.Pool, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.poolStats.pagesAllocPerMinute, prometheus.GaugeValue, s.PagesAllocPerMinute, s.Pool, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.poolStats.pagesDeallocPerMinute, prometheus.GaugeValue, s.PagesDeallocPerMinute, s.Pool, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.poolStats.pagesUnmapPerMinute, prometheus.GaugeValue, s.PagesUnmapPerMinute, s.Pool, s.SerialNumber)

		if len(s.ResettableStatistics) > 0 {
			r := s.ResettableStatistics[0]
			ch <- prometheus.MustNewConstMetric(c.poolStats.avgReadRspTimeMicroseconds, prometheus.GaugeValue, r.AvgReadRspTime, s.Pool, s.SerialNumber)
			ch <- prometheus.MustNewConstMetric(c.poolStats.avgRspTimeMicroseconds, prometheus.GaugeValue, r.AvgRspTime, s.Pool, s.SerialNumber)
			ch <- prometheus.MustNewConstMetric(c.poolStats.avgWriteRspTimeMicroseconds, prometheus.GaugeValue, r.AvgWriteRspTime, s.Pool, s.SerialNumber)
			ch <- prometheus.MustNewConstMetric(c.poolStats.bytesPerSecond, prometheus.GaugeValue, r.BytesPerSecond, s.Pool, s.SerialNumber)
			ch <- prometheus.MustNewConstMetric(c.poolStats.dataReadBytesTotal, prometheus.CounterValue, r.DataReadNumeric, s.Pool, s.SerialNumber)
			ch <- prometheus.MustNewConstMetric(c.poolStats.dataWrittenBytesTotal, prometheus.CounterValue, r.DataWrittenNumeric, s.Pool, s.SerialNumber)
			ch <- prometheus.MustNewConstMetric(c.poolStats.iops, prometheus.GaugeValue, r.IOPS, s.Pool, s.SerialNumber)
			ch <- prometheus.MustNewConstMetric(c.poolStats.readsTotal, prometheus.CounterValue, r.NumberOfReads, s.Pool, s.SerialNumber)
			ch <- prometheus.MustNewConstMetric(c.poolStats.writesTotal, prometheus.CounterValue, r.NumberOfWrites, s.Pool, s.SerialNumber)
		}
	}

	return nil
}
