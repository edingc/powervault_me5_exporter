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

type diskDescs struct {
	avgResponseMs *prometheus.Desc
	health        *prometheus.Desc
	info          *prometheus.Desc
	sizeBytes     *prometheus.Desc
	ssdLifeLeft   *prometheus.Desc
	temperature   *prometheus.Desc
}

func newDiskDescs(label func(string, string, string, ...string) *prometheus.Desc) diskDescs {
	return diskDescs{
		avgResponseMs: label("disk", "average_response_time_microseconds", "Average I/O response time in microseconds.", "serial"),
		health:        label("disk", "health", "Disk health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).", "serial"),
		info:          label("disk", "info", "Disk drive metadata.", "serial", "vendor", "model", "revision", "type", "location", "enclosure_id", "slot"),
		sizeBytes:     label("disk", "size_bytes", "Total size of the disk in bytes.", "serial"),
		ssdLifeLeft:   label("disk", "ssd_life_left_percentage", "For an SSD, this value shows the percentage of disk life remaining.", "serial"),
		temperature:   label("disk", "temperature_celsius", "Disk temperature in Celsius.", "serial"),
	}
}

func (d diskDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.avgResponseMs
	ch <- d.health
	ch <- d.info
	ch <- d.sizeBytes
	ch <- d.ssdLifeLeft
	ch <- d.temperature
}

func (c *ME5Collector) CollectDisks(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		Drives []struct {
			AvgRspTime         float64 `json:"avg-rsp-time"`
			Blocksize          float64 `json:"blocksize"`
			EnclosureID        float64 `json:"enclosure-id"`
			HealthNumeric      float64 `json:"health-numeric"`
			Location           string  `json:"location"`
			Model              string  `json:"model"`
			Revision           string  `json:"revision"`
			SerialNumber       string  `json:"serial-number"`
			SizeNumeric        float64 `json:"size-numeric"`
			Slot               float64 `json:"slot"`
			SSDLifeLeftNumeric float64 `json:"ssd-life-left-numeric"`
			TemperatureNumeric float64 `json:"temperature-numeric"`
			Type               string  `json:"type"`
			Vendor             string  `json:"vendor"`
		} `json:"drives"`
	}

	if err := c.client.Get(ctx, "/show/disks", &resp); err != nil {
		slog.Error("failed to fetch disks from API", "endpoint", "/show/disks", "error", err)
		return err
	}

	if len(resp.Drives) == 0 {
		slog.Warn("API returned success but found no disk drives")
		return nil
	}

	for _, d := range resp.Drives {
		ch <- prometheus.MustNewConstMetric(c.disk.avgResponseMs, prometheus.GaugeValue, d.AvgRspTime, d.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.disk.health, prometheus.GaugeValue, d.HealthNumeric, d.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.disk.info, prometheus.GaugeValue, 1, d.SerialNumber, d.Vendor, d.Model, d.Revision, d.Type, d.Location, strconv.Itoa(int(d.EnclosureID)), strconv.Itoa(int(d.Slot)))
		ch <- prometheus.MustNewConstMetric(c.disk.sizeBytes, prometheus.GaugeValue, d.SizeNumeric*d.Blocksize, d.SerialNumber) // Use reported blocksize
		ch <- prometheus.MustNewConstMetric(c.disk.ssdLifeLeft, prometheus.GaugeValue, d.SSDLifeLeftNumeric, d.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.disk.temperature, prometheus.GaugeValue, d.TemperatureNumeric, d.SerialNumber)
	}

	return nil
}
