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
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCollectDisks(t *testing.T) {
	data, err := os.ReadFile("testdata/disks.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/disks": data},
	}, onlyEnabled(CollectorDisks))

	expected := `
# HELP me5_disk_average_response_time_microseconds Average I/O response time in microseconds.
# TYPE me5_disk_average_response_time_microseconds gauge
me5_disk_average_response_time_microseconds{serial="S0DUMMY0X000001"} 3552
me5_disk_average_response_time_microseconds{serial="S0DUMMY0X000002"} 3476
# HELP me5_disk_health Disk health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).
# TYPE me5_disk_health gauge
me5_disk_health{serial="S0DUMMY0X000001"} 0
me5_disk_health{serial="S0DUMMY0X000002"} 0
# HELP me5_disk_info Disk drive metadata.
# TYPE me5_disk_info gauge
me5_disk_info{enclosure_id="0",location="0.0",model="MZILG7T6HBLAAD3",revision="DSG9",serial="S0DUMMY0X000001",slot="0",type="sSAS",vendor="SAMSUNG"} 1
me5_disk_info{enclosure_id="0",location="0.1",model="MZILG7T6HBLAAD3",revision="DSG9",serial="S0DUMMY0X000002",slot="1",type="sSAS",vendor="SAMSUNG"} 1
# HELP me5_disk_size_bytes Total size of the disk in bytes.
# TYPE me5_disk_size_bytes gauge
me5_disk_size_bytes{serial="S0DUMMY0X000001"} 7.681501126656e+12
me5_disk_size_bytes{serial="S0DUMMY0X000002"} 7.681501126656e+12
# HELP me5_disk_ssd_life_left_percentage For an SSD, this value shows the percentage of disk life remaining.
# TYPE me5_disk_ssd_life_left_percentage gauge
me5_disk_ssd_life_left_percentage{serial="S0DUMMY0X000001"} 100
me5_disk_ssd_life_left_percentage{serial="S0DUMMY0X000002"} 100
# HELP me5_disk_temperature_celsius Disk temperature in Celsius.
# TYPE me5_disk_temperature_celsius gauge
me5_disk_temperature_celsius{serial="S0DUMMY0X000001"} 31
me5_disk_temperature_celsius{serial="S0DUMMY0X000002"} 29
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_disk_average_response_time_microseconds",
		"me5_disk_health",
		"me5_disk_info",
		"me5_disk_size_bytes",
		"me5_disk_ssd_life_left_percentage",
		"me5_disk_temperature_celsius",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectDisks_EmptyResponse(t *testing.T) {
	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/disks": []byte(`{"drives":[]}`)},
	}, onlyEnabled(CollectorDisks))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectDisks(context.Background(), ch); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics for empty response, got %d", len(ch))
	}
}

func TestCollectDisks_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/disks": errors.New("connection refused")},
	}, onlyEnabled(CollectorDisks))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectDisks(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
