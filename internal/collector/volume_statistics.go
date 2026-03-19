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

type volumeStatsDescs struct {
	allocatedPages        *prometheus.Desc
	bytesPerSecond        *prometheus.Desc
	dataReadBytesTotal    *prometheus.Desc
	dataWrittenBytesTotal *prometheus.Desc
	iops                  *prometheus.Desc
	readCacheHitsTotal    *prometheus.Desc
	readCacheMissesTotal  *prometheus.Desc
	readsTotal            *prometheus.Desc
	writeCacheHitsTotal   *prometheus.Desc
	writeCacheMissesTotal *prometheus.Desc
	writesTotal           *prometheus.Desc
}

func newVolumeStatsDescs(label func(string, string, string, ...string) *prometheus.Desc) volumeStatsDescs {
	return volumeStatsDescs{
		allocatedPages:        label("volume", "allocated_pages", "The number of pages allocated to the volume.", "name", "serial"),
		bytesPerSecond:        label("volume", "bytes_per_second", "The data transfer rate, in bytes per second, calculated over the interval since these statistics were last requested or reset.", "name", "serial"),
		dataReadBytesTotal:    label("volume", "data_read_bytes_total", "The amount of data read since these statistics were last reset or since the controller was restarted.", "name", "serial"),
		dataWrittenBytesTotal: label("volume", "data_written_bytes_total", "The amount of data written since these statistics were last reset or since the controller was restarted.", "name", "serial"),
		iops:                  label("volume", "iops", "Input/output operations per second, calculated over the interval since these statistics were last requested or reset.", "name", "serial"),
		readCacheHitsTotal:    label("volume", "read_cache_hits_total", "For the controller that owns the volume, the number of times the block to be read is found in cache.", "name", "serial"),
		readCacheMissesTotal:  label("volume", "read_cache_misses_total", "For the controller that owns the volume, the number of times the block to be read is not found in cache.", "name", "serial"),
		readsTotal:            label("volume", "reads_total", "The number of read operations since these statistics were last reset or since the controller was restarted.", "name", "serial"),
		writeCacheHitsTotal:   label("volume", "write_cache_hits_total", "For the controller that owns the volume, the number of times the block written to is found in cache.", "name", "serial"),
		writeCacheMissesTotal: label("volume", "write_cache_misses_total", "For the controller that owns the volume, the number of times the block written to is not found in cache.", "name", "serial"),
		writesTotal:           label("volume", "writes_total", "The number of write operations since these statistics were last reset or since the controller was restarted.", "name", "serial"),
	}
}

func (d volumeStatsDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.allocatedPages
	ch <- d.bytesPerSecond
	ch <- d.dataReadBytesTotal
	ch <- d.dataWrittenBytesTotal
	ch <- d.iops
	ch <- d.readCacheHitsTotal
	ch <- d.readCacheMissesTotal
	ch <- d.readsTotal
	ch <- d.writeCacheHitsTotal
	ch <- d.writeCacheMissesTotal
	ch <- d.writesTotal
}

func (c *ME5Collector) CollectVolumeStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		VolumeStatistics []struct {
			AllocatedPages     float64 `json:"allocated-pages"`
			BytesPerSecond     float64 `json:"bytes-per-second-numeric"`
			DataReadNumeric    float64 `json:"data-read-numeric"`
			DataWrittenNumeric float64 `json:"data-written-numeric"`
			IOPS               float64 `json:"iops"`
			NumberOfReads      float64 `json:"number-of-reads"`
			NumberOfWrites     float64 `json:"number-of-writes"`
			ReadCacheHits      float64 `json:"read-cache-hits"`
			ReadCacheMisses    float64 `json:"read-cache-misses"`
			SerialNumber       string  `json:"serial-number"`
			VolumeName         string  `json:"volume-name"`
			WriteCacheHits     float64 `json:"write-cache-hits"`
			WriteCacheMisses   float64 `json:"write-cache-misses"`
		} `json:"volume-statistics"`
	}

	if err := c.client.Get(ctx, "/show/volume-statistics", &resp); err != nil {
		slog.Error("failed to fetch volume-statistics from API", "endpoint", "/show/volume-statistics", "error", err)
		return err
	}

	for _, s := range resp.VolumeStatistics {
		ch <- prometheus.MustNewConstMetric(c.volumeStats.allocatedPages, prometheus.GaugeValue, s.AllocatedPages, s.VolumeName, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volumeStats.bytesPerSecond, prometheus.GaugeValue, s.BytesPerSecond, s.VolumeName, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volumeStats.dataReadBytesTotal, prometheus.CounterValue, s.DataReadNumeric, s.VolumeName, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volumeStats.dataWrittenBytesTotal, prometheus.CounterValue, s.DataWrittenNumeric, s.VolumeName, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volumeStats.iops, prometheus.GaugeValue, s.IOPS, s.VolumeName, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volumeStats.readCacheHitsTotal, prometheus.CounterValue, s.ReadCacheHits, s.VolumeName, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volumeStats.readCacheMissesTotal, prometheus.CounterValue, s.ReadCacheMisses, s.VolumeName, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volumeStats.readsTotal, prometheus.CounterValue, s.NumberOfReads, s.VolumeName, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volumeStats.writeCacheHitsTotal, prometheus.CounterValue, s.WriteCacheHits, s.VolumeName, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volumeStats.writeCacheMissesTotal, prometheus.CounterValue, s.WriteCacheMisses, s.VolumeName, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volumeStats.writesTotal, prometheus.CounterValue, s.NumberOfWrites, s.VolumeName, s.SerialNumber)
	}

	return nil
}
