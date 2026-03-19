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

type controllerDateDescs struct {
	dateTimeNumeric *prometheus.Desc
	ntpContactTime  *prometheus.Desc
}

func newControllerDateDescs(label func(string, string, string, ...string) *prometheus.Desc) controllerDateDescs {
	return controllerDateDescs{
		dateTimeNumeric: label("controller_date", "date_time_numeric", "Controller date and time as a Unix timestamp.", "time_zone_region", "ntp_address"),
		ntpContactTime:  label("controller_date", "ntp_contact_time_info", "NTP metadata.", "ntp_state", "ntp_address", "ntp_last_contact_time"),
	}
}

func (d controllerDateDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.dateTimeNumeric
	ch <- d.ntpContactTime
}

func (c *ME5Collector) CollectControllerDate(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		TimeSettings []struct {
			DateTime           string  `json:"date-time"`
			DateTimeNumeric    float64 `json:"date-time-numeric"`
			Meta               string  `json:"meta"`
			NTPAddress         string  `json:"ntp-address"`
			NTPContactTime     string  `json:"ntp-contact-time"`
			NTPState           string  `json:"ntp-state"`
			ObjectName         string  `json:"object-name"`
			TimeZoneDST        string  `json:"time-zone-dst"`
			TimeZoneDSTNumeric float64 `json:"time-zone-dst-numeric"`
			TimeZoneOffset     string  `json:"time-zone-offset"`
			TimeZoneRegion     string  `json:"time-zone-region"`
		} `json:"time-settings-table"`
	}

	if err := c.client.Get(ctx, "/show/controller-date", &resp); err != nil {
		slog.Error("failed to fetch controller date from API", "endpoint", "/show/controller-date", "error", err)
		return err
	}

	for _, ts := range resp.TimeSettings {
		ch <- prometheus.MustNewConstMetric(c.timeSettings.dateTimeNumeric, prometheus.GaugeValue, ts.DateTimeNumeric, ts.TimeZoneRegion, ts.NTPAddress)
		ch <- prometheus.MustNewConstMetric(c.timeSettings.ntpContactTime, prometheus.GaugeValue, 1, ts.NTPState, ts.NTPAddress, ts.NTPContactTime)
	}

	return nil
}
