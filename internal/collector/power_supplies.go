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

type powerSupplyDescs struct {
	health *prometheus.Desc
	info   *prometheus.Desc
	status *prometheus.Desc
}

func newPowerSupplyDescs(label func(string, string, string, ...string) *prometheus.Desc) powerSupplyDescs {
	return powerSupplyDescs{
		health: label("power_supply", "health", "Power supply health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).", "name", "serial"),
		info:   label("power_supply", "info", "Power supply metadata.", "name", "serial", "location", "model", "fw_revision", "part_number", "description"),
		status: label("power_supply", "status", "Power supply status (0=Up, 1=Warning, 2=Error, 3=Not Present, 4=Unknown).", "name", "serial"),
	}
}

func (d powerSupplyDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.health
	ch <- d.info
	ch <- d.status
}

func (c *ME5Collector) CollectPowerSupplies(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		PowerSupplies []struct {
			SerialNumber  string  `json:"serial-number"`
			PartNumber    string  `json:"part-number"`
			Description   string  `json:"description"`
			Name          string  `json:"name"`
			FWRevision    string  `json:"fw-revision"`
			Model         string  `json:"model"`
			Location      string  `json:"location"`
			HealthNumeric float64 `json:"health-numeric"`
			StatusNumeric float64 `json:"status-numeric"`
		} `json:"power-supplies"`
	}

	if err := c.client.Get(ctx, "/show/power-supplies", &resp); err != nil {
		slog.Error("failed to fetch power supplies from API", "endpoint", "/show/power-supplies", "error", err)
		return err
	}

	if len(resp.PowerSupplies) == 0 {
		slog.Warn("API returned success but found no power supply units")
		return nil
	}

	for _, p := range resp.PowerSupplies {
		ch <- prometheus.MustNewConstMetric(c.powerSupply.health, prometheus.GaugeValue, p.HealthNumeric, p.Name, p.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.powerSupply.info, prometheus.GaugeValue, 1, p.Name, p.SerialNumber, p.Location, p.Model, p.FWRevision, p.PartNumber, p.Description)
		ch <- prometheus.MustNewConstMetric(c.powerSupply.status, prometheus.GaugeValue, p.StatusNumeric, p.Name, p.SerialNumber)
	}

	return nil
}
