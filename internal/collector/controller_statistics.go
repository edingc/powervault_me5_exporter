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
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type controllerStatsDescs struct {
	bytesPerSecond        *prometheus.Desc
	cpuLoad               *prometheus.Desc
	dataReadBytesTotal    *prometheus.Desc
	dataWrittenBytesTotal *prometheus.Desc
	iops                  *prometheus.Desc
	numForwardedCmds      *prometheus.Desc
	powerOnHoursTotal     *prometheus.Desc
	powerOnSecondsTotal   *prometheus.Desc
	readCacheHitsTotal    *prometheus.Desc
	readCacheMissesTotal  *prometheus.Desc
	readsTotal            *prometheus.Desc
	writeCacheHitsTotal   *prometheus.Desc
	writeCacheMissesTotal *prometheus.Desc
	writeCacheUsed        *prometheus.Desc
	writesTotal           *prometheus.Desc
}

func newControllerStatsDescs(label func(string, string, string, ...string) *prometheus.Desc) controllerStatsDescs {
	return controllerStatsDescs{
		bytesPerSecond:        label("controller", "bytes_per_second", "The data transfer rate, in bytes per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.", "id"),
		cpuLoad:               label("controller", "cpu_load_percent", "The percentage of time the CPU is busy, from 0 to 100.", "id"),
		dataReadBytesTotal:    label("controller", "data_read_bytes_total", "The amount of data read since these statistics were last reset or since the controller was restarted.", "id"),
		dataWrittenBytesTotal: label("controller", "data_written_bytes_total", "The amount of data written since these statistics were last reset or since the controller was restarted.", "id"),
		iops:                  label("controller", "iops", "The input/output operations per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.", "id"),
		numForwardedCmds:      label("controller", "num_forwarded_cmds_total", "The current count of commands that are being forwarded or are queued to be forwarded to the partner controller for processing. This value will be zero if no commands are being forwarded or are queued to be forwarded.", "id"),
		powerOnHoursTotal:     label("controller", "power_on_hours_total", "The total amount of hours the controller has been powered on in its life time.", "id"),
		powerOnSecondsTotal:   label("controller", "power_on_seconds_total", "The number of seconds since the controller was restarted.", "id"),
		readCacheHitsTotal:    label("controller", "read_cache_hits_total", "For the controller that owns the volume, the number of times the block to be read is found in cache.", "id"),
		readCacheMissesTotal:  label("controller", "read_cache_misses_total", "For the controller that owns the volume, the number of times the block to be read is not found in cache.", "id"),
		readsTotal:            label("controller", "reads_total", "The number of read operations since these statistics were last reset or since the controller was restarted.", "id"),
		writeCacheHitsTotal:   label("controller", "write_cache_hits_total", "For the controller that owns the volume, the number of times the block written to is found in cache.", "id"),
		writeCacheMissesTotal: label("controller", "write_cache_misses_total", "For the controller that owns the volume, the number of times the block written to is not found in cache.", "id"),
		writeCacheUsed:        label("controller", "write_cache_used_percent", "Percentage of write cache in use, from 0 to 100.", "id"),
		writesTotal:           label("controller", "writes_total", "The number of write operations since these statistics were last reset or since the controller was restarted.", "id"),
	}
}

func (d controllerStatsDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.bytesPerSecond
	ch <- d.cpuLoad
	ch <- d.dataReadBytesTotal
	ch <- d.dataWrittenBytesTotal
	ch <- d.iops
	ch <- d.numForwardedCmds
	ch <- d.powerOnHoursTotal
	ch <- d.powerOnSecondsTotal
	ch <- d.readCacheHitsTotal
	ch <- d.readCacheMissesTotal
	ch <- d.readsTotal
	ch <- d.writeCacheHitsTotal
	ch <- d.writeCacheMissesTotal
	ch <- d.writeCacheUsed
	ch <- d.writesTotal
}

func (c *ME5Collector) CollectControllerStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		ControllerStatistics []struct {
			BytesPerSecond     float64 `json:"bytes-per-second-numeric"`
			CPULoad            float64 `json:"cpu-load"`
			ControllerID       string  `json:"controller-id"`
			DataReadNumeric    float64 `json:"data-read-numeric"`
			DataWrittenNumeric float64 `json:"data-written-numeric"`
			DurableID          string  `json:"durable-id"`
			IOPS               float64 `json:"iops"`
			IPAddress          string  `json:"ip-address"`
			NumForwardedCmds   float64 `json:"num-forwarded-cmds"`
			NumberOfReads      float64 `json:"number-of-reads"`
			NumberOfWrites     float64 `json:"number-of-writes"`
			PowerOnTime        float64 `json:"power-on-time"`
			ReadCacheHits      float64 `json:"read-cache-hits"`
			ReadCacheMisses    float64 `json:"read-cache-misses"`
			TotalPowerOnHours  string  `json:"total-power-on-hours"`
			WriteCacheHits     float64 `json:"write-cache-hits"`
			WriteCacheMisses   float64 `json:"write-cache-misses"`
			WriteCacheUsed     float64 `json:"write-cache-used"`
		} `json:"controller-statistics"`
	}

	if err := c.client.Get(ctx, "/show/controller-statistics", &resp); err != nil {
		slog.Error("failed to fetch controller statistics from API", "endpoint", "/show/controller-statistics", "error", err)
		return err
	}

	for _, s := range resp.ControllerStatistics {
		id := s.DurableID
		ch <- prometheus.MustNewConstMetric(c.controllerStats.bytesPerSecond, prometheus.GaugeValue, s.BytesPerSecond, id)
		ch <- prometheus.MustNewConstMetric(c.controllerStats.cpuLoad, prometheus.GaugeValue, s.CPULoad, id)
		ch <- prometheus.MustNewConstMetric(c.controllerStats.dataReadBytesTotal, prometheus.CounterValue, s.DataReadNumeric, id)
		ch <- prometheus.MustNewConstMetric(c.controllerStats.dataWrittenBytesTotal, prometheus.CounterValue, s.DataWrittenNumeric, id)
		ch <- prometheus.MustNewConstMetric(c.controllerStats.iops, prometheus.GaugeValue, s.IOPS, id)
		ch <- prometheus.MustNewConstMetric(c.controllerStats.numForwardedCmds, prometheus.CounterValue, s.NumForwardedCmds, id)
		if hours, err := strconv.ParseFloat(s.TotalPowerOnHours, 64); err == nil {
			ch <- prometheus.MustNewConstMetric(c.controllerStats.powerOnHoursTotal, prometheus.CounterValue, hours, id)
		}
		ch <- prometheus.MustNewConstMetric(c.controllerStats.powerOnSecondsTotal, prometheus.CounterValue, s.PowerOnTime, id)
		ch <- prometheus.MustNewConstMetric(c.controllerStats.readCacheHitsTotal, prometheus.CounterValue, s.ReadCacheHits, id)
		ch <- prometheus.MustNewConstMetric(c.controllerStats.readCacheMissesTotal, prometheus.CounterValue, s.ReadCacheMisses, id)
		ch <- prometheus.MustNewConstMetric(c.controllerStats.readsTotal, prometheus.CounterValue, s.NumberOfReads, id)
		ch <- prometheus.MustNewConstMetric(c.controllerStats.writeCacheHitsTotal, prometheus.CounterValue, s.WriteCacheHits, id)
		ch <- prometheus.MustNewConstMetric(c.controllerStats.writeCacheMissesTotal, prometheus.CounterValue, s.WriteCacheMisses, id)
		ch <- prometheus.MustNewConstMetric(c.controllerStats.writeCacheUsed, prometheus.GaugeValue, s.WriteCacheUsed, id)
		ch <- prometheus.MustNewConstMetric(c.controllerStats.writesTotal, prometheus.CounterValue, s.NumberOfWrites, id)
	}

	return nil
}
