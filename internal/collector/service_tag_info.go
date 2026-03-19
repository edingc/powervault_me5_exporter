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

type serviceTagDescs struct {
	info *prometheus.Desc
}

func newServiceTagDescs(label func(string, string, string, ...string) *prometheus.Desc) serviceTagDescs {
	return serviceTagDescs{
		info: label("service_tag", "info", "Enclosure metadata.", "service_tag", "enclosure_id"),
	}
}

func (d serviceTagDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.info
}

func (c *ME5Collector) CollectServiceTag(ctx context.Context, ch chan<- prometheus.Metric) error {
	var resp struct {
		ServiceTagInfo []struct {
			ServiceTag  string  `json:"service-tag"`
			EnclosureID float64 `json:"enclosure-id"`
		} `json:"service-tag-info"`
	}

	if err := c.client.Get(ctx, "/show/service-tag-info", &resp); err != nil {
		slog.Error("failed to fetch service tag info from API", "endpoint", "/show/service-tag-info", "error", err)
		return err
	}

	for _, st := range resp.ServiceTagInfo {
		ch <- prometheus.MustNewConstMetric(c.serviceTag.info, prometheus.GaugeValue, 1, st.ServiceTag, strconv.Itoa(int(st.EnclosureID)))
	}

	return nil
}
