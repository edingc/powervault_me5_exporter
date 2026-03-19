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

type alertDescs struct {
	bySeverity *prometheus.Desc
}

func newAlertDescs(label func(string, string, string, ...string) *prometheus.Desc) alertDescs {
	return alertDescs{
		bySeverity: label("alerts", "by_severity", "Number of alerts by severity, resolved and acknowledged state.", "severity", "resolved", "acknowledged"),
	}
}

func (d alertDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.bySeverity
}

func (c *ME5Collector) CollectAlerts(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		Alerts []struct {
			Severity     string `json:"severity"`
			Resolved     string `json:"resolved"`
			Acknowledged string `json:"acknowledged"`
		} `json:"alerts"`
	}

	if err := c.client.Get(ctx, "/show/alerts", &resp); err != nil {
		slog.Error("failed to fetch alerts from API", "endpoint", "/show/alerts", "error", err)
		return err
	}

	// Instead of returning individual alerts, return sum of different alert states as metric.
	type key struct{ severity, resolved, acknowledged string }
	counts := make(map[key]float64)

	for _, a := range resp.Alerts {
		counts[key{a.Severity, a.Resolved, a.Acknowledged}]++
	}

	for k, count := range counts {
		ch <- prometheus.MustNewConstMetric(c.alert.bySeverity, prometheus.GaugeValue, count, k.severity, k.resolved, k.acknowledged)
	}

	return nil
}
