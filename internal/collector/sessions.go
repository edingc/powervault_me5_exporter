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

type sessionDescs struct {
	active *prometheus.Desc
	info   *prometheus.Desc
}

func newSessionDescs(label func(string, string, string, ...string) *prometheus.Desc) sessionDescs {
	return sessionDescs{
		active: label("sessions", "active", "Number of active sessions on the storage system."),
		info:   label("sessions", "info", "Active session metadata.", "session_id", "username", "host", "interface"),
	}
}

func (d sessionDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.active
	ch <- d.info
}

func (c *ME5Collector) CollectSessions(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		Sessions []struct {
			Host      string `json:"host"`
			Interface string `json:"interface"`
			SessionID string `json:"sessionId"`
			Username  string `json:"username"`
		} `json:"sessions"`
	}

	if err := c.client.Get(ctx, "/show/sessions", &resp); err != nil {
		slog.Error("failed to fetch sessions from API", "endpoint", "/show/sessions", "error", err)
		return err
	}

	ch <- prometheus.MustNewConstMetric(c.session.active, prometheus.GaugeValue, float64(len(resp.Sessions)))

	for _, s := range resp.Sessions {
		ch <- prometheus.MustNewConstMetric(c.session.info, prometheus.GaugeValue, 1, s.SessionID, s.Username, s.Host, s.Interface)
	}

	return nil
}
