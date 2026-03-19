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

type enclosureDescs struct {
	coolingElementCount *prometheus.Desc
	diskCount           *prometheus.Desc
	drawerHealth        *prometheus.Desc
	drawerStatus        *prometheus.Desc
	expanderHealth      *prometheus.Desc
	expanderInfo        *prometheus.Desc
	health              *prometheus.Desc
	info                *prometheus.Desc
	powerSupplyCount    *prometheus.Desc
	slotCount           *prometheus.Desc
	status              *prometheus.Desc
}

func newEnclosureDescs(label func(string, string, string, ...string) *prometheus.Desc) enclosureDescs {
	encLabels := []string{"enclosure_id"}
	drwLabels := []string{"enclosure_id", "drawer_id"}
	expLabels := []string{"enclosure_id", "drawer_id", "name", "location"}

	return enclosureDescs{
		coolingElementCount: label("enclosure", "cooling_element_count", "Number of fan units in the enclosure.", encLabels...),
		diskCount:           label("enclosure", "disk_count", "Number of disk slots (not installed disks) in the enclosure.", encLabels...),
		drawerHealth:        label("enclosure", "drawer_health", "Drawer health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).", drwLabels...),
		drawerStatus:        label("enclosure", "drawer_status", "Drawer status (0=Unsupported, 1=OK, 2=Critical, 3=Warning, 4=Unrecoverable, 5=Not Installed, 6=Unknown, 7=Unavailable).", drwLabels...),
		expanderHealth:      label("enclosure", "expander_health", "Expander health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).", expLabels...),
		expanderInfo:        label("enclosure", "expander_info", "Expander metadata.", append(expLabels, "revision")...),
		health:              label("enclosure", "health", "Enclosure health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).", encLabels...),
		info:                label("enclosure", "info", "Enclosure metadata.", "enclosure_id", "vendor", "model", "midplane_serial", "wwn", "revision"),
		powerSupplyCount:    label("enclosure", "power_supply_count", "Number of power supplies in the enclosure.", encLabels...),
		slotCount:           label("enclosure", "slot_count", "Number of disk slots in this enclosure.", encLabels...),
		status:              label("enclosure", "status", "Enclosure status (0=Unsupported, 1=OK, 2=Critical, 3=Warning, 4=Unrecoverable, 5=Not Installed, 6=Unknown, 7=Unavailable).", encLabels...),
	}
}

func (d enclosureDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.coolingElementCount
	ch <- d.diskCount
	ch <- d.drawerHealth
	ch <- d.drawerStatus
	ch <- d.expanderHealth
	ch <- d.expanderInfo
	ch <- d.health
	ch <- d.info
	ch <- d.powerSupplyCount
	ch <- d.slotCount
	ch <- d.status
}

func (c *ME5Collector) CollectEnclosures(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		Enclosures []struct {
			EnclosureID             float64 `json:"enclosure-id"`
			EnclosureWWN            string  `json:"enclosure-wwn"`
			HealthNumeric           float64 `json:"health-numeric"`
			MidplaneSerialNumber    string  `json:"midplane-serial-number"`
			Model                   string  `json:"model"`
			NumberOfCoolingElements float64 `json:"number-of-coolings-elements"`
			NumberOfDisks           float64 `json:"number-of-disks"`
			NumberOfPowerSupplies   float64 `json:"number-of-power-supplies"`
			Revision                string  `json:"revision"`
			Slots                   float64 `json:"slots"`
			StatusNumeric           float64 `json:"status-numeric"`
			Vendor                  string  `json:"vendor"`
			Drawers                 []struct {
				DrawerID      float64 `json:"drawer-id"`
				HealthNumeric float64 `json:"health-numeric"`
				StatusNumeric float64 `json:"status-numeric"`
				Sideplanes    []struct {
					Expanders []struct {
						FWRevision    string  `json:"fw-revision"`
						HealthNumeric float64 `json:"health-numeric"`
						Location      string  `json:"location"`
						Name          string  `json:"name"`
						StatusNumeric float64 `json:"status-numeric"`
					} `json:"expanders"`
				} `json:"sideplanes"`
			} `json:"drawers"`
		} `json:"enclosures"`
	}

	if err := c.client.Get(ctx, "/show/enclosures", &resp); err != nil {
		slog.Error("failed to fetch enclosures from API", "endpoint", "/show/enclosures", "error", err)
		return err
	}

	for _, e := range resp.Enclosures {
		encID := strconv.Itoa(int(e.EnclosureID))
		ch <- prometheus.MustNewConstMetric(c.enclosure.coolingElementCount, prometheus.GaugeValue, e.NumberOfCoolingElements, encID)
		ch <- prometheus.MustNewConstMetric(c.enclosure.diskCount, prometheus.GaugeValue, e.NumberOfDisks, encID)
		ch <- prometheus.MustNewConstMetric(c.enclosure.health, prometheus.GaugeValue, e.HealthNumeric, encID)
		ch <- prometheus.MustNewConstMetric(c.enclosure.info, prometheus.GaugeValue, 1, encID, e.Vendor, e.Model, e.MidplaneSerialNumber, e.EnclosureWWN, e.Revision)
		ch <- prometheus.MustNewConstMetric(c.enclosure.powerSupplyCount, prometheus.GaugeValue, e.NumberOfPowerSupplies, encID)
		ch <- prometheus.MustNewConstMetric(c.enclosure.slotCount, prometheus.GaugeValue, e.Slots, encID)
		ch <- prometheus.MustNewConstMetric(c.enclosure.status, prometheus.GaugeValue, e.StatusNumeric, encID)

		for _, d := range e.Drawers {
			drwID := strconv.Itoa(int(d.DrawerID))
			ch <- prometheus.MustNewConstMetric(c.enclosure.drawerHealth, prometheus.GaugeValue, d.HealthNumeric, encID, drwID)
			ch <- prometheus.MustNewConstMetric(c.enclosure.drawerStatus, prometheus.GaugeValue, d.StatusNumeric, encID, drwID)

			for _, s := range d.Sideplanes {
				for _, exp := range s.Expanders {
					expLv := []string{encID, drwID, exp.Name, exp.Location}
					ch <- prometheus.MustNewConstMetric(c.enclosure.expanderHealth, prometheus.GaugeValue, exp.HealthNumeric, expLv...)
					ch <- prometheus.MustNewConstMetric(c.enclosure.expanderInfo, prometheus.GaugeValue, 1, append(expLv, exp.FWRevision)...)
				}
			}
		}
	}

	return nil
}
