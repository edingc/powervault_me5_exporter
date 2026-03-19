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
	"regexp"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var sensorValuePattern = regexp.MustCompile(`[0-9]*\.?[0-9]+`)

type sensorDescs struct {
	status *prometheus.Desc
	value  *prometheus.Desc
}

func newSensorDescs(label func(string, string, string, ...string) *prometheus.Desc) sensorDescs {
	return sensorDescs{
		status: label("sensor", "status", "Sensor status (0=Unsupported, 1=OK, 2=Critical, 3=Warning, 4=Unrecoverable, 5=Not Installed, 6=Unknown, 7=Unavailable).", "name", "type", "container", "controller", "id"),
		value:  label("sensor", "value", "Sensor value.", "name", "type", "container", "controller", "id"),
	}
}

func (d sensorDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.status
	ch <- d.value
}

func (c *ME5Collector) CollectSensors(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		Sensors []struct {
			DurableID     string  `json:"durable-id"`
			ControllerID  string  `json:"controller-id"`
			SensorName    string  `json:"sensor-name"`
			Value         string  `json:"value"`
			StatusNumeric float64 `json:"status-numeric"`
			Container     string  `json:"container"`
			SensorType    string  `json:"sensor-type"`
		} `json:"sensors"`
	}

	if err := c.client.Get(ctx, "/show/sensor-status", &resp); err != nil {
		slog.Error("failed to fetch sensor status from API", "endpoint", "/show/sensor-status", "error", err)
		return err
	}

	if len(resp.Sensors) == 0 {
		slog.Warn("API returned success but found no sensor units")
		return nil
	}

	for _, s := range resp.Sensors {
		ch <- prometheus.MustNewConstMetric(c.sensor.status, prometheus.GaugeValue, s.StatusNumeric, s.SensorName, s.SensorType, s.Container, s.ControllerID, s.DurableID)

		if match := sensorValuePattern.FindString(s.Value); match != "" {
			if val, err := strconv.ParseFloat(match, 64); err == nil {
				ch <- prometheus.MustNewConstMetric(c.sensor.value, prometheus.GaugeValue, val, s.SensorName, s.SensorType, s.Container, s.ControllerID, s.DurableID)
			} else {
				slog.Debug("failed to parse extracted sensor value", "sensor", s.SensorName, "raw", s.Value, "extracted", match, "error", err)
			}
		}
	}

	return nil
}
