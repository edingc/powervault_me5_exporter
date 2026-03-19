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

type firmwareBundleDescs struct {
	health *prometheus.Desc
	info   *prometheus.Desc
}

func newFirmwareBundleDescs(label func(string, string, string, ...string) *prometheus.Desc) firmwareBundleDescs {
	return firmwareBundleDescs{
		health: label("firmware_bundle", "health", "Firmware bundle health (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).", "version"),
		info:   label("firmware_bundle", "info", "Firmware bundle metadata.", "version", "build_date", "status_name"),
	}
}

func (d firmwareBundleDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.health
	ch <- d.info
}

func (c *ME5Collector) CollectFirmwareBundles(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		Bundles []struct {
			BuildDate     string  `json:"build-date"`
			BundleVersion string  `json:"bundle-version"`
			HealthNumeric float64 `json:"health-numeric"`
			Status        string  `json:"status"`
		} `json:"firmware-bundles"`
	}

	if err := c.client.Get(ctx, "/show/firmware-bundles", &resp); err != nil {
		slog.Error("failed to fetch firmware bundles from API", "endpoint", "/show/firmware-bundles", "error", err)
		return err
	}

	for _, b := range resp.Bundles {
		ch <- prometheus.MustNewConstMetric(c.firmwareBundle.info, prometheus.GaugeValue, 1, b.BundleVersion, b.BuildDate, b.Status)
		ch <- prometheus.MustNewConstMetric(c.firmwareBundle.health, prometheus.GaugeValue, b.HealthNumeric, b.BundleVersion)
	}

	return nil
}
