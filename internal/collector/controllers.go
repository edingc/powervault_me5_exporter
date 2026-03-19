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

type controllerDescs struct {
	disks            *prometheus.Desc
	failoverStatus   *prometheus.Desc
	health           *prometheus.Desc
	info             *prometheus.Desc
	redundancyStatus *prometheus.Desc
	status           *prometheus.Desc
	storagePools     *prometheus.Desc
	virtualDisks     *prometheus.Desc
}

func newControllerDescs(label func(string, string, string, ...string) *prometheus.Desc) controllerDescs {
	return controllerDescs{
		disks:            label("controller", "disks", "Number of disks in the storage system.", "id", "ip"),
		failoverStatus:   label("controller", "failover_status", "Controller failover status. (0=No, 1=Yes)", "id", "ip"),
		health:           label("controller", "health", "Controller health status. (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A)", "id", "ip"),
		info:             label("controller", "info", "Controller metadata.", "id", "vendor", "model", "serial", "part_number", "revision", "description", "mfg_date", "hw_version", "cpld_version", "ip", "mac", "wwn", "position", "redundancy_mode", "drive_bus_type", "host_ports", "drive_channels", "cache_memory_size", "system_memory_size"),
		redundancyStatus: label("controller", "redundancy_status", "Controller redundancy status. (0=Operational - Not Redundant, 2=Redundant, 4=Down, 5=Unknown)", "id", "ip"),
		status:           label("controller", "status", "Controller status. (0=Operational, 1=Down, 2=Not Installed)", "id", "ip"),
		storagePools:     label("controller", "storage_pools", "Number of virtual pools in the storage system.", "id", "ip"),
		virtualDisks:     label("controller", "virtual_disks", "Number of disk groups in the storage system.", "id", "ip"),
	}
}

func (d controllerDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.disks
	ch <- d.failoverStatus
	ch <- d.health
	ch <- d.info
	ch <- d.redundancyStatus
	ch <- d.status
	ch <- d.storagePools
	ch <- d.virtualDisks
}

func (c *ME5Collector) CollectControllers(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		Controllers []struct {
			CPLDVersion             string  `json:"cpld-version"`
			CacheMemorySize         float64 `json:"cache-memory-size"`
			ControllerID            string  `json:"controller-id"`
			Description             string  `json:"description"`
			Disks                   float64 `json:"disks"`
			DriveBusType            string  `json:"drive-bus-type"`
			DriveChannels           float64 `json:"drive-channels"`
			FailedOverNumeric       float64 `json:"failed-over-numeric"`
			HardwareVersion         string  `json:"hardware-version"`
			HealthNumeric           float64 `json:"health-numeric"`
			HostPorts               float64 `json:"host-ports"`
			IPAddress               string  `json:"ip-address"`
			MACAddress              string  `json:"mac-address"`
			MfgDate                 string  `json:"mfg-date"`
			Model                   string  `json:"model"`
			NodeWWN                 string  `json:"node-wwn"`
			NumberOfStoragePools    float64 `json:"number-of-storage-pools"`
			PartNumber              string  `json:"part-number"`
			Position                string  `json:"position"`
			RedundancyMode          string  `json:"redundancy-mode"`
			RedundancyStatusNumeric float64 `json:"redundancy-status-numeric"`
			Revision                string  `json:"revision"`
			SerialNumber            string  `json:"serial-number"`
			StatusNumeric           float64 `json:"status-numeric"`
			SystemMemorySize        float64 `json:"system-memory-size"`
			Vendor                  string  `json:"vendor"`
			VirtualDisks            float64 `json:"virtual-disks"`
		} `json:"controllers"`
	}

	if err := c.client.Get(ctx, "/show/controllers", &resp); err != nil {
		slog.Error("failed to fetch controllers from API", "endpoint", "/show/controllers", "error", err)
		return err
	}

	for _, ctrl := range resp.Controllers {
		id := ctrl.ControllerID
		ip := ctrl.IPAddress

		ch <- prometheus.MustNewConstMetric(c.controller.disks, prometheus.GaugeValue, ctrl.Disks, id, ip)
		ch <- prometheus.MustNewConstMetric(c.controller.failoverStatus, prometheus.GaugeValue, ctrl.FailedOverNumeric, id, ip)
		ch <- prometheus.MustNewConstMetric(c.controller.health, prometheus.GaugeValue, ctrl.HealthNumeric, id, ip)
		ch <- prometheus.MustNewConstMetric(c.controller.info, prometheus.GaugeValue, 1, id, ctrl.Vendor, ctrl.Model, ctrl.SerialNumber, ctrl.PartNumber, ctrl.Revision, ctrl.Description, ctrl.MfgDate, ctrl.HardwareVersion, ctrl.CPLDVersion, ip, ctrl.MACAddress, ctrl.NodeWWN, ctrl.Position, ctrl.RedundancyMode, ctrl.DriveBusType, strconv.Itoa(int(ctrl.HostPorts)), strconv.Itoa(int(ctrl.DriveChannels)), strconv.Itoa(int(ctrl.CacheMemorySize)), strconv.Itoa(int(ctrl.SystemMemorySize)))
		ch <- prometheus.MustNewConstMetric(c.controller.redundancyStatus, prometheus.GaugeValue, ctrl.RedundancyStatusNumeric, id, ip)
		ch <- prometheus.MustNewConstMetric(c.controller.status, prometheus.GaugeValue, ctrl.StatusNumeric, id, ip)
		ch <- prometheus.MustNewConstMetric(c.controller.storagePools, prometheus.GaugeValue, ctrl.NumberOfStoragePools, id, ip)
		ch <- prometheus.MustNewConstMetric(c.controller.virtualDisks, prometheus.GaugeValue, ctrl.VirtualDisks, id, ip)
	}

	return nil
}
