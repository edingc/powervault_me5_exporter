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

type diskGroupDescs struct {
	allocatedPages *prometheus.Desc
	availablePages *prometheus.Desc
	diskCount      *prometheus.Desc
	freespaceBytes *prometheus.Desc
	health         *prometheus.Desc
	info           *prometheus.Desc
	sizeBytes      *prometheus.Desc
	spareCount     *prometheus.Desc
	status         *prometheus.Desc
}

func newDiskGroupDescs(label func(string, string, string, ...string) *prometheus.Desc) diskGroupDescs {
	commonLabels := []string{"name", "serial"}

	return diskGroupDescs{
		allocatedPages: label("disk_group", "allocated_pages", "For a virtual pool, the number of 4 MB pages that are currently in use. For a linear pool, 0.", commonLabels...),
		availablePages: label("disk_group", "available_pages", "For a virtual pool, the number of 4 MB pages that are still available to be allocated. For a linear pool, 0.", commonLabels...),
		diskCount:      label("disk_group", "disk_count", "Number of disks in the disk group.", commonLabels...),
		freespaceBytes: label("disk_group", "freespace_bytes", "The amount of free space in the disk group, in bytes.", commonLabels...),
		health:         label("disk_group", "health", "Disk group health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).", commonLabels...),
		info:           label("disk_group", "info", "Disk group metadata.", "name", "serial", "pool", "raidtype", "storage_tier", "owner"),
		sizeBytes:      label("disk_group", "size_bytes", "Disk group capacity, in bytes.", commonLabels...),
		spareCount:     label("disk_group", "spare_count", "For a linear disk group, the number of spares assigned to the disk group. For a virtual disk group, 0.", commonLabels...),
		status:         label("disk_group", "status", "Disk group status (0=FTOL, 1=FTDN, 2=CRIT, 3=OFFL, 4=QTCR, 5=QTOF, 6=QTDN, 7=STOP, 8=MSNG, 9=DMGD, 250=UP, other=UNKN).", commonLabels...),
	}
}

func (d diskGroupDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.allocatedPages
	ch <- d.availablePages
	ch <- d.diskCount
	ch <- d.freespaceBytes
	ch <- d.health
	ch <- d.info
	ch <- d.sizeBytes
	ch <- d.spareCount
	ch <- d.status
}

func (c *ME5Collector) CollectDiskGroups(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		DiskGroups []struct {
			AllocatedPages       float64 `json:"allocated-pages"`
			AvailablePages       float64 `json:"available-pages"`
			Blocksize            float64 `json:"blocksize"`
			CurrentJobCompletion string  `json:"current-job-completion"`
			DiskCount            float64 `json:"diskcount"`
			FreespaceNumeric     float64 `json:"freespace-numeric"`
			HealthNumeric        float64 `json:"health-numeric"`
			Name                 string  `json:"name"`
			Owner                string  `json:"owner"`
			Pool                 string  `json:"pool"`
			Raidtype             string  `json:"raidtype"`
			SerialNumber         string  `json:"serial-number"`
			SizeNumeric          float64 `json:"size-numeric"`
			SpareCount           float64 `json:"sparecount"`
			StatusNumeric        float64 `json:"status-numeric"`
			StorageTier          string  `json:"storage-tier"`
		} `json:"disk-groups"`
	}

	if err := c.client.Get(ctx, "/show/disk-groups", &resp); err != nil {
		slog.Error("failed to fetch disk groups from API", "endpoint", "/show/disk-groups", "error", err)
		return err
	}

	for _, dg := range resp.DiskGroups {
		name := dg.Name
		sn := dg.SerialNumber

		ch <- prometheus.MustNewConstMetric(c.diskGroup.allocatedPages, prometheus.GaugeValue, dg.AllocatedPages, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroup.availablePages, prometheus.GaugeValue, dg.AvailablePages, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroup.diskCount, prometheus.GaugeValue, dg.DiskCount, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroup.freespaceBytes, prometheus.GaugeValue, dg.FreespaceNumeric*dg.Blocksize, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroup.health, prometheus.GaugeValue, dg.HealthNumeric, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroup.info, prometheus.GaugeValue, 1, name, sn, dg.Pool, dg.Raidtype, dg.StorageTier, dg.Owner)
		ch <- prometheus.MustNewConstMetric(c.diskGroup.sizeBytes, prometheus.GaugeValue, dg.SizeNumeric*dg.Blocksize, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroup.spareCount, prometheus.GaugeValue, dg.SpareCount, name, sn)
		ch <- prometheus.MustNewConstMetric(c.diskGroup.status, prometheus.GaugeValue, dg.StatusNumeric, name, sn)
	}

	return nil
}
