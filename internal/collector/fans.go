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

type fanDescs struct {
	health *prometheus.Desc
	speed  *prometheus.Desc
	status *prometheus.Desc
}

func newFanDescs(label func(string, string, string, ...string) *prometheus.Desc) fanDescs {
	return fanDescs{
		health: label("fan", "health", "Fan health status. (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A)", "name", "location"),
		speed:  label("fan", "speed", "Fan speed (revolutions per minute).", "name", "location"),
		status: label("fan", "status", "Fan unit status. (0=Up, 1=Error, 2=Off, 3=Missing)", "name", "location"),
	}
}

func (d fanDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.health
	ch <- d.speed
	ch <- d.status
}

func (c *ME5Collector) CollectFans(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		Fans []struct {
			Name          string  `json:"name"`
			Location      string  `json:"location"`
			StatusNumeric float64 `json:"status-numeric"`
			Speed         float64 `json:"speed"`
			HealthNumeric float64 `json:"health-numeric"`
		} `json:"fan"`
	}

	if err := c.client.Get(ctx, "/show/fans", &resp); err != nil {
		slog.Error("failed to fetch fans from API", "endpoint", "/show/fans", "error", err)
		return err
	}

	if len(resp.Fans) == 0 {
		slog.Warn("API returned success but found no fan units")
		return nil
	}

	for _, fan := range resp.Fans {
		ch <- prometheus.MustNewConstMetric(c.fan.health, prometheus.GaugeValue, fan.HealthNumeric, fan.Name, fan.Location)
		ch <- prometheus.MustNewConstMetric(c.fan.status, prometheus.GaugeValue, fan.StatusNumeric, fan.Name, fan.Location)
		ch <- prometheus.MustNewConstMetric(c.fan.speed, prometheus.GaugeValue, fan.Speed, fan.Name, fan.Location)
	}

	return nil
}
