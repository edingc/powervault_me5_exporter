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

type snapshotSpaceDescs struct {
	allocatedBytes *prometheus.Desc
	limitBytes     *prometheus.Desc
}

func newSnapshotSpaceDescs(label func(string, string, string, ...string) *prometheus.Desc) snapshotSpaceDescs {
	return snapshotSpaceDescs{
		allocatedBytes: label("snapshot_space", "allocated_bytes", "Snapshot space currently allocated in bytes.", "pool", "serial"),
		limitBytes:     label("snapshot_space", "limit_bytes", "Snapshot space limit in bytes.", "pool", "serial"),
	}
}

func (d snapshotSpaceDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.allocatedBytes
	ch <- d.limitBytes
}

func (c *ME5Collector) CollectSnapshotSpace(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		SnapSpace []struct {
			AllocatedSizeNumeric float64 `json:"allocated-size-numeric"`
			Pool                 string  `json:"pool"`
			SerialNumber         string  `json:"serial-number"`
			SnapLimitSizeNumeric float64 `json:"snap-limit-size-numeric"`
		} `json:"snap-space"`
	}

	if err := c.client.Get(ctx, "/show/snapshot-space", &resp); err != nil {
		slog.Error("failed to fetch snapshot-space from API", "endpoint", "/show/snapshot-space", "error", err)
		return err
	}

	for _, s := range resp.SnapSpace {
		ch <- prometheus.MustNewConstMetric(c.snapshotSpace.allocatedBytes, prometheus.GaugeValue, s.AllocatedSizeNumeric*512, s.Pool, s.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.snapshotSpace.limitBytes, prometheus.GaugeValue, s.SnapLimitSizeNumeric*512, s.Pool, s.SerialNumber)
	}

	return nil
}
