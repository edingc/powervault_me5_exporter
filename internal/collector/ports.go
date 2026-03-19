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

type portDescs struct {
	actualSpeed      *prometheus.Desc
	health           *prometheus.Desc
	info             *prometheus.Desc
	sasActiveLanes   *prometheus.Desc
	sasExpectedLanes *prometheus.Desc
	status           *prometheus.Desc
}

func newPortDescs(label func(string, string, string, ...string) *prometheus.Desc) portDescs {
	return portDescs{
		actualSpeed:      label("port", "actual_speed_numeric", "Actual link speed (0=1Gb, 1=2Gb, 2=4Gb, 3=Port Disconnected, 6=6Gb, 7=8Gb, 8=10Mb, 9=100Mb, 11=12Gb, 12=16Gb).", "port", "controller"),
		health:           label("port", "health", "Port health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).", "port", "controller"),
		info:             label("port", "info", "Port metadata.", "port", "controller", "type", "media", "target_id", "configured_speed"),
		sasActiveLanes:   label("port", "sas_active_lanes", "Number of active SAS lanes.", "port", "controller"),
		sasExpectedLanes: label("port", "sas_expected_lanes", "Number of expected SAS lanes.", "port", "controller"),
		status:           label("port", "status", "Port status (0=Up, 1=Warning, 2=Error, 3=Not Present, 6=Disconnected).", "port", "controller"),
	}
}

func (d portDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.actualSpeed
	ch <- d.health
	ch <- d.info
	ch <- d.sasActiveLanes
	ch <- d.sasExpectedLanes
	ch <- d.status
}

func (c *ME5Collector) CollectPorts(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		Ports []struct {
			ActualSpeedNumeric float64 `json:"actual-speed-numeric"`
			ConfiguredSpeed    string  `json:"configured-speed"`
			Controller         string  `json:"controller"`
			HealthNumeric      float64 `json:"health-numeric"`
			Media              string  `json:"media"`
			Port               string  `json:"port"`
			PortType           string  `json:"port-type"`
			SASPort            []struct {
				SASActiveLanes   float64 `json:"sas-active-lanes"`
				SASLanesExpected float64 `json:"sas-lanes-expected"`
			} `json:"sas-port"`
			StatusNumeric float64 `json:"status-numeric"`
			TargetID      string  `json:"target-id"`
		} `json:"port"`
	}

	if err := c.client.Get(ctx, "/show/ports", &resp); err != nil {
		slog.Error("failed to fetch ports from API", "endpoint", "/show/ports", "error", err)
		return err
	}

	for _, p := range resp.Ports {
		ch <- prometheus.MustNewConstMetric(c.port.actualSpeed, prometheus.GaugeValue, p.ActualSpeedNumeric, p.Port, p.Controller)
		ch <- prometheus.MustNewConstMetric(c.port.health, prometheus.GaugeValue, p.HealthNumeric, p.Port, p.Controller)
		ch <- prometheus.MustNewConstMetric(c.port.info, prometheus.GaugeValue, 1, p.Port, p.Controller, p.PortType, p.Media, p.TargetID, p.ConfiguredSpeed)
		ch <- prometheus.MustNewConstMetric(c.port.status, prometheus.GaugeValue, p.StatusNumeric, p.Port, p.Controller)

		var activeLanes, expectedLanes float64
		if len(p.SASPort) > 0 {
			activeLanes = p.SASPort[0].SASActiveLanes
			expectedLanes = p.SASPort[0].SASLanesExpected
		}
		ch <- prometheus.MustNewConstMetric(c.port.sasActiveLanes, prometheus.GaugeValue, activeLanes, p.Port, p.Controller)
		ch <- prometheus.MustNewConstMetric(c.port.sasExpectedLanes, prometheus.GaugeValue, expectedLanes, p.Port, p.Controller)
	}

	return nil
}
