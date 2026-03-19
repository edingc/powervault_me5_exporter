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

type diskStatsDescs struct {
	badBlocksTotal                *prometheus.Desc
	blockReassignsTotal           *prometheus.Desc
	bytesPerSecond                *prometheus.Desc
	dataReadBytesTotal            *prometheus.Desc
	dataWrittenBytesTotal         *prometheus.Desc
	ioTimeoutCountTotal           *prometheus.Desc
	iops                          *prometheus.Desc
	lifetimeDataReadBytesTotal    *prometheus.Desc
	lifetimeDataWrittenBytesTotal *prometheus.Desc
	mediaErrorsTotal              *prometheus.Desc
	noResponseCountTotal          *prometheus.Desc
	nonmediaErrorsTotal           *prometheus.Desc
	powerOnHoursTotal             *prometheus.Desc
	queueDepth                    *prometheus.Desc
	readsTotal                    *prometheus.Desc
	smartCount                    *prometheus.Desc
	spinupRetryCountTotal         *prometheus.Desc
	writesTotal                   *prometheus.Desc
}

func newDiskStatsDescs(label func(string, string, string, ...string) *prometheus.Desc) diskStatsDescs {
	commonLabels := []string{"location", "serial"}
	pathLabels := append(commonLabels, "path")

	return diskStatsDescs{
		badBlocksTotal:                label("disk", "bad_blocks_total", "Total bad block count by controller path.", pathLabels...),
		blockReassignsTotal:           label("disk", "block_reassigns_total", "Total block reassignment count by controller path.", pathLabels...),
		bytesPerSecond:                label("disk", "bytes_per_second", "The data transfer rate, in bytes per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.", commonLabels...),
		dataReadBytesTotal:            label("disk", "data_read_bytes_total", "Amount of data read since these statistics were last reset or since the controller was restarted, in bytes.", commonLabels...),
		dataWrittenBytesTotal:         label("disk", "data_written_bytes_total", "Amount of data written since these statistics were last reset or since the controller was restarted, in bytes.", commonLabels...),
		ioTimeoutCountTotal:           label("disk", "io_timeout_count_total", "Total I/O timeout count by controller path.", pathLabels...),
		iops:                          label("disk", "iops", "Input/output operations per second, calculated over the interval since these statistics were last requested or reset. This value will be zero if it has not been requested or reset since a controller restart.", commonLabels...),
		lifetimeDataReadBytesTotal:    label("disk", "lifetime_data_read_bytes_total", "The amount of data read from the disk in its lifetime, in bytes.", commonLabels...),
		lifetimeDataWrittenBytesTotal: label("disk", "lifetime_data_written_bytes_total", "The amount of data written to the disk in its lifetime, in bytes.", commonLabels...),
		mediaErrorsTotal:              label("disk", "media_errors_total", "Total media error count by controller path.", pathLabels...),
		noResponseCountTotal:          label("disk", "no_response_count_total", "Total no-response count by controller path.", pathLabels...),
		nonmediaErrorsTotal:           label("disk", "nonmedia_errors_total", "Total non-media error count by controller path.", pathLabels...),
		powerOnHoursTotal:             label("disk", "power_on_hours_total", "The total number of hours that the disk has been powered on since it was manufactured. This value is stored in disk metadata and is updated in 30- minute increments.", commonLabels...),
		queueDepth:                    label("disk", "queue_depth", "Number of pending I/O operations currently being serviced.", commonLabels...),
		readsTotal:                    label("disk", "reads_total", "Number of read operations since these statistics were last reset or since the controller was restarted.", commonLabels...),
		smartCount:                    label("disk", "smart_count", "SMART event count by controller path.", pathLabels...),
		spinupRetryCountTotal:         label("disk", "spinup_retry_count_total", "Total spinup retry count by controller path.", pathLabels...),
		writesTotal:                   label("disk", "writes_total", "Number of write operations since these statistics were last reset or since the controller was restarted.", commonLabels...),
	}
}

func (d diskStatsDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.badBlocksTotal
	ch <- d.blockReassignsTotal
	ch <- d.bytesPerSecond
	ch <- d.dataReadBytesTotal
	ch <- d.dataWrittenBytesTotal
	ch <- d.ioTimeoutCountTotal
	ch <- d.iops
	ch <- d.lifetimeDataReadBytesTotal
	ch <- d.lifetimeDataWrittenBytesTotal
	ch <- d.mediaErrorsTotal
	ch <- d.noResponseCountTotal
	ch <- d.nonmediaErrorsTotal
	ch <- d.powerOnHoursTotal
	ch <- d.queueDepth
	ch <- d.readsTotal
	ch <- d.smartCount
	ch <- d.spinupRetryCountTotal
	ch <- d.writesTotal
}

func (c *ME5Collector) CollectDiskStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		DiskStatistics []struct {
			BytesPerSecondNumeric      float64 `json:"bytes-per-second-numeric"`
			DataReadNumeric            float64 `json:"data-read-numeric"`
			DataWrittenNumeric         float64 `json:"data-written-numeric"`
			IOPS                       float64 `json:"iops"`
			IOTimeoutCount1            float64 `json:"io-timeout-count-1"`
			IOTimeoutCount2            float64 `json:"io-timeout-count-2"`
			LifetimeDataReadNumeric    float64 `json:"lifetime-data-read-numeric"`
			LifetimeDataWrittenNumeric float64 `json:"lifetime-data-written-numeric"`
			Location                   string  `json:"location"`
			NoResponseCount1           float64 `json:"no-response-count-1"`
			NoResponseCount2           float64 `json:"no-response-count-2"`
			NumberOfBadBlocks1         float64 `json:"number-of-bad-blocks-1"`
			NumberOfBadBlocks2         float64 `json:"number-of-bad-blocks-2"`
			NumberOfBlockReassigns1    float64 `json:"number-of-block-reassigns-1"`
			NumberOfBlockReassigns2    float64 `json:"number-of-block-reassigns-2"`
			NumberOfMediaErrors1       float64 `json:"number-of-media-errors-1"`
			NumberOfMediaErrors2       float64 `json:"number-of-media-errors-2"`
			NumberOfNonmediaErrors1    float64 `json:"number-of-nonmedia-errors-1"`
			NumberOfNonmediaErrors2    float64 `json:"number-of-nonmedia-errors-2"`
			NumberOfReads              float64 `json:"number-of-reads"`
			NumberOfWrites             float64 `json:"number-of-writes"`
			PowerOnHours               float64 `json:"power-on-hours"`
			QueueDepth                 float64 `json:"queue-depth"`
			SerialNumber               string  `json:"serial-number"`
			SmartCount1                float64 `json:"smart-count-1"`
			SmartCount2                float64 `json:"smart-count-2"`
			SpinupRetryCount1          float64 `json:"spinup-retry-count-1"`
			SpinupRetryCount2          float64 `json:"spinup-retry-count-2"`
		} `json:"disk-statistics"`
	}

	if err := c.client.Get(ctx, "/show/disk-statistics", &resp); err != nil {
		slog.Error("failed to fetch disk statistics from API", "endpoint", "/show/disk-statistics", "error", err)
		return err
	}

	for _, s := range resp.DiskStatistics {
		loc := s.Location
		sn := s.SerialNumber

		ch <- prometheus.MustNewConstMetric(c.diskStats.bytesPerSecond, prometheus.GaugeValue, s.BytesPerSecondNumeric, loc, sn)
		ch <- prometheus.MustNewConstMetric(c.diskStats.dataReadBytesTotal, prometheus.CounterValue, s.DataReadNumeric, loc, sn)
		ch <- prometheus.MustNewConstMetric(c.diskStats.dataWrittenBytesTotal, prometheus.CounterValue, s.DataWrittenNumeric, loc, sn)
		ch <- prometheus.MustNewConstMetric(c.diskStats.iops, prometheus.GaugeValue, s.IOPS, loc, sn)
		ch <- prometheus.MustNewConstMetric(c.diskStats.lifetimeDataReadBytesTotal, prometheus.CounterValue, s.LifetimeDataReadNumeric, loc, sn)
		ch <- prometheus.MustNewConstMetric(c.diskStats.lifetimeDataWrittenBytesTotal, prometheus.CounterValue, s.LifetimeDataWrittenNumeric, loc, sn)
		ch <- prometheus.MustNewConstMetric(c.diskStats.powerOnHoursTotal, prometheus.CounterValue, s.PowerOnHours, loc, sn)
		ch <- prometheus.MustNewConstMetric(c.diskStats.queueDepth, prometheus.GaugeValue, s.QueueDepth, loc, sn)
		ch <- prometheus.MustNewConstMetric(c.diskStats.readsTotal, prometheus.CounterValue, s.NumberOfReads, loc, sn)
		ch <- prometheus.MustNewConstMetric(c.diskStats.writesTotal, prometheus.CounterValue, s.NumberOfWrites, loc, sn)

		// Path-specific metrics
		paths := []struct {
			id string
			bt float64
			rc float64
			it float64
			nr float64
			sc float64
			sr float64
			me float64
			ne float64
		}{
			{"1", s.NumberOfBadBlocks1, s.NumberOfBlockReassigns1, s.IOTimeoutCount1, s.NoResponseCount1, s.SmartCount1, s.SpinupRetryCount1, s.NumberOfMediaErrors1, s.NumberOfNonmediaErrors1},
			{"2", s.NumberOfBadBlocks2, s.NumberOfBlockReassigns2, s.IOTimeoutCount2, s.NoResponseCount2, s.SmartCount2, s.SpinupRetryCount2, s.NumberOfMediaErrors2, s.NumberOfNonmediaErrors2},
		}

		for _, p := range paths {
			ch <- prometheus.MustNewConstMetric(c.diskStats.badBlocksTotal, prometheus.CounterValue, p.bt, loc, sn, p.id)
			ch <- prometheus.MustNewConstMetric(c.diskStats.blockReassignsTotal, prometheus.CounterValue, p.rc, loc, sn, p.id)
			ch <- prometheus.MustNewConstMetric(c.diskStats.ioTimeoutCountTotal, prometheus.CounterValue, p.it, loc, sn, p.id)
			ch <- prometheus.MustNewConstMetric(c.diskStats.noResponseCountTotal, prometheus.CounterValue, p.nr, loc, sn, p.id)
			ch <- prometheus.MustNewConstMetric(c.diskStats.smartCount, prometheus.GaugeValue, p.sc, loc, sn, p.id)
			ch <- prometheus.MustNewConstMetric(c.diskStats.spinupRetryCountTotal, prometheus.CounterValue, p.sr, loc, sn, p.id)
			ch <- prometheus.MustNewConstMetric(c.diskStats.mediaErrorsTotal, prometheus.CounterValue, p.me, loc, sn, p.id)
			ch <- prometheus.MustNewConstMetric(c.diskStats.nonmediaErrorsTotal, prometheus.CounterValue, p.ne, loc, sn, p.id)
		}
	}

	return nil
}
