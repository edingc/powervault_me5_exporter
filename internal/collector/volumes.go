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

type volumeDescs struct {
	allocatedSizeBytes *prometheus.Desc
	health             *prometheus.Desc
	info               *prometheus.Desc
	metadataSizeBytes  *prometheus.Desc
	sizeBytes          *prometheus.Desc
	totalSizeBytes     *prometheus.Desc
}

func newVolumeDescs(label func(string, string, string, ...string) *prometheus.Desc) volumeDescs {
	return volumeDescs{
		allocatedSizeBytes: label("volume", "allocated_bytes", "The amount of space currently allocated to a virtual volume, or the total size of a linear volume, in bytes.", "name", "serial"),
		health:             label("volume", "health", "Volume health status. (0=OK, 1=Degraded, 2=Fault, 3=Unknown)", "name", "serial"),
		info:               label("volume", "info", "Volume information.", "name", "serial", "virtual_disk_name", "storage_pool_name", "storage_type", "usage_type", "owner", "volume_type", "volume_class", "tier_affinity", "is_snapshot", "parent_volume", "creation_date_time", "description", "wwn", "allowed_storage_tiers", "raidtype", "volume_group"),
		metadataSizeBytes:  label("volume", "metadata_bytes", "Amount of pool metadata currently being used by the volume, in bytes.", "name", "serial"),
		sizeBytes:          label("volume", "size_bytes", "Total volume capacity, in bytes.", "name", "serial"),
		totalSizeBytes:     label("volume", "total_bytes", "The total size of the volume, in bytes.", "name", "serial"),
	}
}

func (d volumeDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.allocatedSizeBytes
	ch <- d.health
	ch <- d.info
	ch <- d.metadataSizeBytes
	ch <- d.sizeBytes
	ch <- d.totalSizeBytes
}

func (c *ME5Collector) CollectVolumes(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		Volumes []struct {
			AllocatedSizeNumeric float64 `json:"allocated-size-numeric"`
			AllowedStorageTiers  string  `json:"allowed-storage-tiers"`
			Blocksize            float64 `json:"blocksize"`
			CreationDateTime     string  `json:"creation-date-time"`
			Health               string  `json:"health"`
			HealthNumeric        float64 `json:"health-numeric"`
			MetadataInUseNumeric float64 `json:"metadata-in-use-numeric"`
			Owner                string  `json:"owner"`
			Raidtype             string  `json:"raidtype"`
			SerialNumber         string  `json:"serial-number"`
			SizeNumeric          float64 `json:"size-numeric"`
			Snapshot             string  `json:"snapshot"`
			StoragePoolName      string  `json:"storage-pool-name"`
			StorageType          string  `json:"storage-type"`
			TierAffinity         string  `json:"tier-affinity"`
			TotalSizeNumeric     float64 `json:"total-size-numeric"`
			VirtualDiskName      string  `json:"virtual-disk-name"`
			VolumeClass          string  `json:"volume-class"`
			VolumeDescription    string  `json:"volume-description"`
			VolumeGroup          string  `json:"volume-group"`
			VolumeName           string  `json:"volume-name"`
			VolumeParent         string  `json:"volume-parent"`
			VolumeType           string  `json:"volume-type"`
			VolumeUsageType      string  `json:"volume-usage-type"`
			WWN                  string  `json:"wwn"`
		} `json:"volumes"`
	}

	if err := c.client.Get(ctx, "/show/volumes", &resp); err != nil {
		slog.Error("failed to fetch volumes from API", "endpoint", "/show/volumes", "error", err)
		return err
	}

	for _, v := range resp.Volumes {
		ch <- prometheus.MustNewConstMetric(c.volume.allocatedSizeBytes, prometheus.GaugeValue, v.AllocatedSizeNumeric*v.Blocksize, v.VolumeName, v.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volume.health, prometheus.GaugeValue, v.HealthNumeric, v.VolumeName, v.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volume.info, prometheus.GaugeValue, 1, v.VolumeName, v.SerialNumber, v.VirtualDiskName, v.StoragePoolName, v.StorageType, v.VolumeUsageType, v.Owner, v.VolumeType, v.VolumeClass, v.TierAffinity, v.Snapshot, v.VolumeParent, v.CreationDateTime, v.VolumeDescription, v.WWN, v.AllowedStorageTiers, v.Raidtype, v.VolumeGroup)
		ch <- prometheus.MustNewConstMetric(c.volume.metadataSizeBytes, prometheus.GaugeValue, v.MetadataInUseNumeric*v.Blocksize, v.VolumeName, v.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volume.sizeBytes, prometheus.GaugeValue, v.SizeNumeric*v.Blocksize, v.VolumeName, v.SerialNumber)
		ch <- prometheus.MustNewConstMetric(c.volume.totalSizeBytes, prometheus.GaugeValue, v.TotalSizeNumeric*v.Blocksize, v.VolumeName, v.SerialNumber)
	}

	return nil
}
