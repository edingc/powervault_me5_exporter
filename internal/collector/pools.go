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

type poolDescs struct {
	allocatedPages *prometheus.Desc
	availableBytes *prometheus.Desc
	availablePages *prometheus.Desc
	health         *prometheus.Desc
	info           *prometheus.Desc
	overcommitted  *prometheus.Desc
	rfcSizeBytes   *prometheus.Desc
	sizeBytes      *prometheus.Desc
	volumes        *prometheus.Desc
}

func newPoolDescs(label func(string, string, string, ...string) *prometheus.Desc) poolDescs {
	return poolDescs{
		allocatedPages: label("pool", "allocated_pages", "For a virtual pool, the number of 4 MB pages that are currently in use. For a linear pool, 0.", "pool", "serial"),
		availableBytes: label("pool", "available_bytes", "The available capacity in the pool, in bytes.", "pool", "serial"),
		availablePages: label("pool", "available_pages", "For a virtual pool, the number of 4 MB pages that are still available to be allocated. For a linear pool, 0.", "pool", "serial"),
		health:         label("pool", "health", "Pool health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).", "pool", "serial"),
		info:           label("pool", "info", "Pool metadata.", "pool", "serial", "storage_type", "owner", "preferred_owner", "sector_format"),
		overcommitted:  label("pool", "overcommitted", "Whether the pool is overcommitted (0=No, 1=Yes).", "pool", "serial"),
		rfcSizeBytes:   label("pool", "rfc_size_bytes", "The total size of the read cache in the pool, in bytes.", "pool", "serial"),
		sizeBytes:      label("pool", "size_bytes", "The total capacity of the pool, in bytes.", "pool", "serial"),
		volumes:        label("pool", "volumes", "The number of volumes in the pool.", "pool", "serial"),
	}
}

func (d poolDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.allocatedPages
	ch <- d.availableBytes
	ch <- d.availablePages
	ch <- d.health
	ch <- d.info
	ch <- d.overcommitted
	ch <- d.rfcSizeBytes
	ch <- d.sizeBytes
	ch <- d.volumes
}

func (c *ME5Collector) CollectPools(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		Pools []struct {
			AllocatedPages       float64 `json:"allocated-pages"`
			AvailablePages       float64 `json:"available-pages"`
			Blocksize            float64 `json:"blocksize"`
			HealthNumeric        float64 `json:"health-numeric"`
			Name                 string  `json:"name"`
			OverCommittedNumeric float64 `json:"over-committed-numeric"`
			Owner                string  `json:"owner"`
			PoolSectorFormat     string  `json:"pool-sector-format"`
			PreferredOwner       string  `json:"preferred-owner"`
			SerialNumber         string  `json:"serial-number"`
			StorageType          string  `json:"storage-type"`
			TotalAvailNumeric    float64 `json:"total-avail-numeric"`
			TotalRfcSizeNumeric  float64 `json:"total-rfc-size-numeric"`
			TotalSizeNumeric     float64 `json:"total-size-numeric"`
			Volumes              float64 `json:"volumes"`
		} `json:"pools"`
	}

	if err := c.client.Get(ctx, "/show/pools", &resp); err != nil {
		slog.Error("failed to fetch pools from API", "endpoint", "/show/pools", "error", err)
		return err
	}

	for _, p := range resp.Pools {
		ch <- prometheus.MustNewConstMetric(c.pool.allocatedPages, prometheus.GaugeValue, p.AllocatedPages, p.Name, p.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.pool.availableBytes, prometheus.GaugeValue, p.TotalAvailNumeric*p.Blocksize, p.Name, p.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.pool.availablePages, prometheus.GaugeValue, p.AvailablePages, p.Name, p.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.pool.health, prometheus.GaugeValue, p.HealthNumeric, p.Name, p.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.pool.info, prometheus.GaugeValue, 1, p.Name, p.SerialNumber, p.StorageType, p.Owner, p.PreferredOwner, p.PoolSectorFormat)
		ch <- prometheus.MustNewConstMetric(c.pool.overcommitted, prometheus.GaugeValue, p.OverCommittedNumeric, p.Name, p.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.pool.rfcSizeBytes, prometheus.GaugeValue, p.TotalRfcSizeNumeric*p.Blocksize, p.Name, p.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.pool.sizeBytes, prometheus.GaugeValue, p.TotalSizeNumeric*p.Blocksize, p.Name, p.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.pool.volumes, prometheus.GaugeValue, p.Volumes, p.Name, p.SerialNumber)
	}

	return nil
}
