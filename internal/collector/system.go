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

type systemDescs struct {
	controllerAStatus *prometheus.Desc
	controllerBStatus *prometheus.Desc
	enclosureCount    *prometheus.Desc
	fdeSecurityStatus *prometheus.Desc
	health            *prometheus.Desc
	info              *prometheus.Desc
	redundancyMode    *prometheus.Desc
	redundancyStatus  *prometheus.Desc
}

func newSystemDescs(label func(string, string, string, ...string) *prometheus.Desc) systemDescs {
	return systemDescs{
		controllerAStatus: label("system", "controller_a_status", "System controller A status (0=Operational, 1=Down, 2=Not Installed).", "serial"),
		controllerBStatus: label("system", "controller_b_status", "System controller B status (0=Operational, 1=Down, 2=Not Installed).", "serial"),
		enclosureCount:    label("system", "enclosure_count", "The number of enclosures in the system."),
		fdeSecurityStatus: label("system", "fde_security_status", "Full Disk Encryption security status (1=Unsecured, 2=Secured, 3=Secured - Lock Ready, 4=Secure - Locked)."),
		health:            label("system", "health", "Overall system health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A)."),
		info:              label("system", "info", "System metadata.", "system_name", "system_contact", "system_location", "system_information", "midplane_serial", "vendor_name", "product_brand", "product_id"),
		redundancyMode:    label("system", "redundancy_mode", "System redundancy mode (8=Active-Active ULP, 9=Single Controller, 10=Failed Over, 11=Down)."),
		redundancyStatus:  label("system", "redundancy_status", "System redundancy status (0=Operational but not redundant, 2=Redundant, 4=Down, 5=Unknown)."),
	}
}

func (d systemDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.controllerAStatus
	ch <- d.controllerBStatus
	ch <- d.enclosureCount
	ch <- d.fdeSecurityStatus
	ch <- d.health
	ch <- d.info
	ch <- d.redundancyMode
	ch <- d.redundancyStatus
}

func (c *ME5Collector) CollectSystem(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		System []struct {
			EnclosureCount           float64 `json:"enclosure-count"`
			FDESecurityStatusNumeric float64 `json:"fde-security-status-numeric"`
			HealthNumeric            float64 `json:"health-numeric"`
			MidplaneSN               string  `json:"midplane-serial-number"`
			ProductBrand             string  `json:"product-brand"`
			ProductID                string  `json:"product-id"`
			Redundancy               []struct {
				ControllerAStatus float64 `json:"controller-a-status-numeric"`
				ControllerASN     string  `json:"controller-a-serial-number"`
				ControllerBStatus float64 `json:"controller-b-status-numeric"`
				ControllerBSN     string  `json:"controller-b-serial-number"`
				RedundancyMode    float64 `json:"redundancy-mode-numeric"`
				RedundancyStatus  float64 `json:"redundancy-status-numeric"`
			} `json:"redundancy"`
			SysContact  string `json:"system-contact"`
			SysInfo     string `json:"system-information"`
			SysLocation string `json:"system-location"`
			SysName     string `json:"system-name"`
			Vendor      string `json:"vendor-name"`
		} `json:"system"`
	}

	if err := c.client.Get(ctx, "/show/system", &resp); err != nil {
		slog.Error("failed to fetch system from API", "endpoint", "/show/system", "error", err)
		return err
	}

	for _, s := range resp.System {
		ch <- prometheus.MustNewConstMetric(c.system.enclosureCount, prometheus.GaugeValue, s.EnclosureCount)
		ch <- prometheus.MustNewConstMetric(c.system.fdeSecurityStatus, prometheus.GaugeValue, s.FDESecurityStatusNumeric)
		ch <- prometheus.MustNewConstMetric(c.system.health, prometheus.GaugeValue, s.HealthNumeric)
		ch <- prometheus.MustNewConstMetric(c.system.info, prometheus.GaugeValue, 1, s.SysName, s.SysContact, s.SysLocation, s.SysInfo, s.MidplaneSN, s.Vendor, s.ProductBrand, s.ProductID)

		if len(s.Redundancy) > 0 {
			r := s.Redundancy[0]
			ch <- prometheus.MustNewConstMetric(c.system.controllerAStatus, prometheus.GaugeValue, r.ControllerAStatus, r.ControllerASN)
			ch <- prometheus.MustNewConstMetric(c.system.controllerBStatus, prometheus.GaugeValue, r.ControllerBStatus, r.ControllerBSN)
			ch <- prometheus.MustNewConstMetric(c.system.redundancyMode, prometheus.GaugeValue, r.RedundancyMode)
			ch <- prometheus.MustNewConstMetric(c.system.redundancyStatus, prometheus.GaugeValue, r.RedundancyStatus)
		}
	}

	return nil
}
