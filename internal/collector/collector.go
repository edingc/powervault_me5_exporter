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
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector name constants for use with enable/disable flags.
const (
	CollectorAlerts          = "alerts"
	CollectorControllerDate  = "controller_date"
	CollectorControllerStats = "controller_statistics"
	CollectorControllers     = "controllers"
	CollectorDiskGroupStats  = "disk_group_statistics"
	CollectorDiskGroups      = "disk_groups"
	CollectorDiskStats       = "disk_statistics"
	CollectorDisks           = "disks"
	CollectorEnclosures      = "enclosures"
	CollectorEvents          = "events"
	CollectorExpanderPorts   = "expander_ports"
	CollectorFans            = "fans"
	CollectorFirmwareBundles = "firmware_bundles"
	CollectorFRUs            = "frus"
	CollectorHostPortStats   = "host_port_statistics"
	CollectorPoolStats       = "pool_statistics"
	CollectorPools           = "pools"
	CollectorPorts           = "ports"
	CollectorPowerSupplies   = "power_supplies"
	CollectorSensors         = "sensors"
	CollectorServiceTag      = "service_tag"
	CollectorSessions        = "sessions"
	CollectorSnapshotSpace   = "snapshot_space"
	CollectorSystem          = "system"
	CollectorVolumeStats     = "volume_statistics"
	CollectorVolumes         = "volumes"
)

// AllCollectors maps each collector name to whether it is enabled by default.
var AllCollectors = map[string]bool{
	CollectorAlerts:          true,
	CollectorControllerDate:  true,
	CollectorControllerStats: true,
	CollectorControllers:     true,
	CollectorDiskGroupStats:  true,
	CollectorDiskGroups:      true,
	CollectorDiskStats:       true,
	CollectorDisks:           true,
	CollectorEnclosures:      true,
	CollectorEvents:          false, // Events do not mean as much as alerts, recommend disabled
	CollectorFans:            true,
	CollectorFirmwareBundles: true,
	CollectorFRUs:            true,
	CollectorHostPortStats:   true,
	CollectorPoolStats:       true,
	CollectorPools:           true,
	CollectorPorts:           true,
	CollectorPowerSupplies:   true,
	CollectorSensors:         true,
	CollectorServiceTag:      true,
	CollectorSessions:        true,
	CollectorSnapshotSpace:   true,
	CollectorSystem:          true,
	CollectorVolumeStats:     true,
	CollectorVolumes:         true,
}

// CollectorHelp maps each collector name to a short description for CLI flags.
var CollectorHelp = map[string]string{
	CollectorAlerts:          "Collect alert counts by severity.",
	CollectorControllerDate:  "Collect controller date/time and NTP metrics.",
	CollectorControllerStats: "Collect controller I/O statistics.",
	CollectorControllers:     "Collect controller metrics.",
	CollectorDiskGroupStats:  "Collect disk group I/O statistics.",
	CollectorDiskGroups:      "Collect disk group metrics.",
	CollectorDiskStats:       "Collect per-disk I/O statistics.",
	CollectorDisks:           "Collect disk metrics.",
	CollectorEnclosures:      "Collect enclosure metrics.",
	CollectorEvents:          "Collect event counts by severity.",
	CollectorFans:            "Collect fan metrics.",
	CollectorFirmwareBundles: "Collect firmware bundle info.",
	CollectorFRUs:            "Collect FRU status metrics.",
	CollectorHostPortStats:   "Collect host port I/O statistics.",
	CollectorPoolStats:       "Collect pool I/O statistics.",
	CollectorPools:           "Collect pool metrics.",
	CollectorPorts:           "Collect port metrics.",
	CollectorPowerSupplies:   "Collect power supply metrics.",
	CollectorSensors:         "Collect sensor status metrics.",
	CollectorServiceTag:      "Collect enclosure service tag info.",
	CollectorSessions:        "Collect active session count.",
	CollectorSnapshotSpace:   "Collect snapshot space metrics.",
	CollectorSystem:          "Collect system-level metrics.",
	CollectorVolumeStats:     "Collect volume I/O statistics.",
	CollectorVolumes:         "Collect volume metrics.",
}

// APIClient is the interface the collector uses to fetch data from the ME5.
type APIClient interface {
	Get(ctx context.Context, path string, dest any) error
}

// ME5Collector implements prometheus.Collector for the Dell PowerVault ME5.
type ME5Collector struct {
	client        APIClient
	enabled       map[string]bool
	scrapeTimeout time.Duration
	concurrency   int

	// Descriptors (Alphabetized)
	alert           alertDescs
	controller      controllerDescs
	controllerStats controllerStatsDescs
	disk            diskDescs
	diskGroup       diskGroupDescs
	diskGroupStats  diskGroupStatsDescs
	diskStats       diskStatsDescs
	enclosure       enclosureDescs
	event           eventDescs
	fan             fanDescs
	firmwareBundle  firmwareBundleDescs
	fru             fruDescs
	hostPortStats   hostPortStatsDescs
	pool            poolDescs
	poolStats       poolStatsDescs
	port            portDescs
	powerSupply     powerSupplyDescs
	sensor          sensorDescs
	serviceTag      serviceTagDescs
	session         sessionDescs
	snapshotSpace   snapshotSpaceDescs
	system          systemDescs
	timeSettings    controllerDateDescs
	volume          volumeDescs
	volumeStats     volumeStatsDescs

	// Scrape meta
	scrapeDurationSeconds *prometheus.Desc
	scrapeSuccess         *prometheus.Desc
}

func NewME5Collector(client APIClient, enabled map[string]bool, scrapeTimeout time.Duration, concurrency int) *ME5Collector {
	const ns = "me5"
	label := func(subsystem, name, help string, variableLabels ...string) *prometheus.Desc {
		return prometheus.NewDesc(prometheus.BuildFQName(ns, subsystem, name), help, variableLabels, nil)
	}

	return &ME5Collector{
		client:        client,
		enabled:       enabled,
		scrapeTimeout: scrapeTimeout,
		concurrency:   concurrency,

		alert:           newAlertDescs(label),
		controller:      newControllerDescs(label),
		controllerStats: newControllerStatsDescs(label),
		disk:            newDiskDescs(label),
		diskGroup:       newDiskGroupDescs(label),
		diskGroupStats:  newDiskGroupStatsDescs(label),
		diskStats:       newDiskStatsDescs(label),
		enclosure:       newEnclosureDescs(label),
		event:           newEventDescs(label),
		fan:             newFanDescs(label),
		firmwareBundle:  newFirmwareBundleDescs(label),
		fru:             newFRUDescs(label),
		hostPortStats:   newHostPortStatsDescs(label),
		pool:            newPoolDescs(label),
		poolStats:       newPoolStatsDescs(label),
		port:            newPortDescs(label),
		powerSupply:     newPowerSupplyDescs(label),
		sensor:          newSensorDescs(label),
		serviceTag:      newServiceTagDescs(label),
		session:         newSessionDescs(label),
		snapshotSpace:   newSnapshotSpaceDescs(label),
		system:          newSystemDescs(label),
		timeSettings:    newControllerDateDescs(label),
		volume:          newVolumeDescs(label),
		volumeStats:     newVolumeStatsDescs(label),

		scrapeDurationSeconds: label("scrape", "duration_seconds", "Duration of the ME5 API scrape in seconds"),
		scrapeSuccess:         label("scrape", "success", "Whether the last ME5 API scrape succeeded (1=success, 0=failure)"),
	}
}

func (c *ME5Collector) Describe(ch chan<- *prometheus.Desc) {
	c.alert.describe(ch)
	c.controller.describe(ch)
	c.controllerStats.describe(ch)
	c.disk.describe(ch)
	c.diskGroup.describe(ch)
	c.diskGroupStats.describe(ch)
	c.diskStats.describe(ch)
	c.enclosure.describe(ch)
	c.event.describe(ch)
	c.fan.describe(ch)
	c.firmwareBundle.describe(ch)
	c.fru.describe(ch)
	c.hostPortStats.describe(ch)
	c.pool.describe(ch)
	c.poolStats.describe(ch)
	c.port.describe(ch)
	c.powerSupply.describe(ch)
	c.sensor.describe(ch)
	c.serviceTag.describe(ch)
	c.session.describe(ch)
	c.snapshotSpace.describe(ch)
	c.system.describe(ch)
	c.timeSettings.describe(ch)
	c.volume.describe(ch)
	c.volumeStats.describe(ch)
	ch <- c.scrapeDurationSeconds
	ch <- c.scrapeSuccess
}

func (c *ME5Collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	slog.Debug("Scrape started")
	// Set a deadline for the entire scrape to ensure we don't hang indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), c.scrapeTimeout)
	defer cancel()

	success := 1.0
	if err := c.collect(ctx, ch); err != nil {
		slog.Error("ME5 scrape failed", "error", err)
		success = 0
	}

	elapsed := time.Since(start)
	slog.Debug("Scrape finished", "duration", elapsed.Round(time.Millisecond), "success", success == 1.0)
	ch <- prometheus.MustNewConstMetric(c.scrapeSuccess, prometheus.GaugeValue, success)
	ch <- prometheus.MustNewConstMetric(c.scrapeDurationSeconds, prometheus.GaugeValue, elapsed.Seconds())
}

func (c *ME5Collector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	collectors := []struct {
		name    string
		collect func(context.Context, chan<- prometheus.Metric) error
	}{
		{CollectorAlerts, c.CollectAlerts},
		{CollectorControllerDate, c.CollectControllerDate},
		{CollectorControllers, c.CollectControllers},
		{CollectorControllerStats, c.CollectControllerStats},
		{CollectorDiskGroupStats, c.CollectDiskGroupStats},
		{CollectorDiskGroups, c.CollectDiskGroups},
		{CollectorDiskStats, c.CollectDiskStats},
		{CollectorDisks, c.CollectDisks},
		{CollectorEnclosures, c.CollectEnclosures},
		{CollectorEvents, c.CollectEvents},
		{CollectorFans, c.CollectFans},
		{CollectorFirmwareBundles, c.CollectFirmwareBundles},
		{CollectorFRUs, c.CollectFRUs},
		{CollectorHostPortStats, c.CollectHostPortStats},
		{CollectorPoolStats, c.CollectPoolStats},
		{CollectorPools, c.CollectPools},
		{CollectorPorts, c.CollectPorts},
		{CollectorPowerSupplies, c.CollectPowerSupplies},
		{CollectorSensors, c.CollectSensors},
		{CollectorServiceTag, c.CollectServiceTag},
		{CollectorSessions, c.CollectSessions},
		{CollectorSnapshotSpace, c.CollectSnapshotSpace},
		{CollectorSystem, c.CollectSystem},
		{CollectorVolumeStats, c.CollectVolumeStats},
		{CollectorVolumes, c.CollectVolumes},
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		hasError bool
		sem      = make(chan struct{}, c.concurrency)
	)

	for _, col := range collectors {
		if enabled, ok := c.enabled[col.name]; ok && !enabled {
			continue
		}

		wg.Add(1)
		go func(name string, fn func(context.Context, chan<- prometheus.Metric) error) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			colStart := time.Now()
			slog.Debug("Sub-collector starting", "collector", name)
			if err := fn(ctx, ch); err != nil {
				slog.Warn("Sub-collector failed", "collector", name, "error", err)
				mu.Lock()
				hasError = true
				mu.Unlock()
			}
			slog.Debug("Sub-collector finished", "collector", name, "duration", time.Since(colStart).Round(time.Millisecond))
		}(col.name, col.collect)
	}

	wg.Wait()

	if hasError {
		return fmt.Errorf("one or more sub-collectors failed")
	}
	return nil
}
