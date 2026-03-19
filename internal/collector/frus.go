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
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type fruDescs struct {
	info   *prometheus.Desc
	status *prometheus.Desc
}

func newFRUDescs(label func(string, string, string, ...string) *prometheus.Desc) fruDescs {
	return fruDescs{
		info:   label("fru", "info", "FRU metadata.", "name", "serial", "part_number", "revision", "description", "location", "enclosure_id", "mfg_date"),
		status: label("fru", "status", "FRU status (0=Invalid Data, 1=Fault, 2= Absent, 3=Power Off, 4=OK).", "name", "serial"),
	}
}

func (d fruDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.info
	ch <- d.status
}

func (c *ME5Collector) CollectFRUs(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		FRUs []struct {
			Description      string  `json:"description"`
			EnclosureID      float64 `json:"enclosure-id"`
			FRULocation      string  `json:"fru-location"`
			FRUStatusNumeric float64 `json:"fru-status-numeric"`
			MfgDate          string  `json:"mfg-date"`
			Name             string  `json:"name"`
			PartNumber       string  `json:"part-number"`
			Revision         string  `json:"revision"`
			SerialNumber     string  `json:"serial-number"`
		} `json:"enclosure-fru"`
	}

	if err := c.client.Get(ctx, "/show/frus", &resp); err != nil {
		slog.Error("failed to fetch FRUs from API", "endpoint", "/show/frus", "error", err)
		return err
	}

	for _, f := range resp.FRUs {
		ch <- prometheus.MustNewConstMetric(c.fru.info, prometheus.GaugeValue, 1, f.Name, f.SerialNumber, f.PartNumber, f.Revision, f.Description, f.FRULocation, strconv.Itoa(int(f.EnclosureID)), f.MfgDate)
		ch <- prometheus.MustNewConstMetric(c.fru.status, prometheus.GaugeValue, f.FRUStatusNumeric, f.Name, f.SerialNumber)
	}

	return nil
}
