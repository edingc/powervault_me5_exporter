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

func TestCollectDiskGroups(t *testing.T) {
	data, err := os.ReadFile("testdata/disk-groups.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/disk-groups": data},
	}, onlyEnabled(CollectorDiskGroups))

	expected := `
# HELP me5_disk_group_allocated_pages For a virtual pool, the number of 4 MB pages that are currently in use. For a linear pool, 0.
# TYPE me5_disk_group_allocated_pages gauge
me5_disk_group_allocated_pages{name="dgA01",serial="00c0fffa39f700001295066800000000"} 2.3956275e+07
# HELP me5_disk_group_available_pages For a virtual pool, the number of 4 MB pages that are still available to be allocated. For a linear pool, 0.
# TYPE me5_disk_group_available_pages gauge
me5_disk_group_available_pages{name="dgA01",serial="00c0fffa39f700001295066800000000"} 1.8318541e+07
# HELP me5_disk_group_disk_count Number of disks in the disk group.
# TYPE me5_disk_group_disk_count gauge
me5_disk_group_disk_count{name="dgA01",serial="00c0fffa39f700001295066800000000"} 28
# HELP me5_disk_group_freespace_bytes The amount of free space in the disk group, in bytes.
# TYPE me5_disk_group_freespace_bytes gauge
me5_disk_group_freespace_bytes{name="dgA01",serial="00c0fffa39f700001295066800000000"} 7.6833529790464e+13
# HELP me5_disk_group_health Disk group health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).
# TYPE me5_disk_group_health gauge
me5_disk_group_health{name="dgA01",serial="00c0fffa39f700001295066800000000"} 0
# HELP me5_disk_group_info Disk group metadata.
# TYPE me5_disk_group_info gauge
me5_disk_group_info{name="dgA01",owner="A",pool="A",raidtype="ADAPT",serial="00c0fffa39f700001295066800000000",storage_tier="Performance"} 1
# HELP me5_disk_group_size_bytes Disk group capacity, in bytes.
# TYPE me5_disk_group_size_bytes gauge
me5_disk_group_size_bytes{name="dgA01",serial="00c0fffa39f700001295066800000000"} 1.77313429848064e+14
# HELP me5_disk_group_spare_count For a linear disk group, the number of spares assigned to the disk group. For a virtual disk group, 0.
# TYPE me5_disk_group_spare_count gauge
me5_disk_group_spare_count{name="dgA01",serial="00c0fffa39f700001295066800000000"} 0
# HELP me5_disk_group_status Disk group status (0=FTOL, 1=FTDN, 2=CRIT, 3=OFFL, 4=QTCR, 5=QTOF, 6=QTDN, 7=STOP, 8=MSNG, 9=DMGD, 250=UP, other=UNKN).
# TYPE me5_disk_group_status gauge
me5_disk_group_status{name="dgA01",serial="00c0fffa39f700001295066800000000"} 0
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_disk_group_allocated_pages",
		"me5_disk_group_available_pages",
		"me5_disk_group_disk_count",
		"me5_disk_group_freespace_bytes",
		"me5_disk_group_health",
		"me5_disk_group_info",
		"me5_disk_group_size_bytes",
		"me5_disk_group_spare_count",
		"me5_disk_group_status",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectDiskGroups_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/disk-groups": errors.New("connection refused")},
	}, onlyEnabled(CollectorDiskGroups))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectDiskGroups(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
