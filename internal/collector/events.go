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

type eventDescs struct {
	bySeverity *prometheus.Desc
}

func newEventDescs(label func(string, string, string, ...string) *prometheus.Desc) eventDescs {
	return eventDescs{
		bySeverity: label("events", "by_severity", "Number of events grouped by severity.", "severity"),
	}
}

func (d eventDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.bySeverity
}

func (c *ME5Collector) CollectEvents(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		Events []struct {
			Severity string `json:"severity"`
		} `json:"events"`
	}

	if err := c.client.Get(ctx, "/show/events", &resp); err != nil {
		slog.Error("failed to fetch events from API", "endpoint", "/show/events", "error", err)
		return err
	}

	// Return the count of each severity as a metric.
	counts := make(map[string]float64)
	for _, e := range resp.Events {
		counts[e.Severity]++
	}

	for severity, count := range counts {
		ch <- prometheus.MustNewConstMetric(c.event.bySeverity, prometheus.GaugeValue, count, severity)
	}

	return nil
}
