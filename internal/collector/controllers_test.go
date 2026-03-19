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

func TestCollectControllers(t *testing.T) {
	data, err := os.ReadFile("testdata/controllers.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/controllers": data},
	}, onlyEnabled(CollectorControllers))

	expected := `
# HELP me5_controller_disks Number of disks in the storage system.
# TYPE me5_controller_disks gauge
me5_controller_disks{id="A",ip="192.0.2.1"} 28
me5_controller_disks{id="B",ip="192.0.2.2"} 28
# HELP me5_controller_failover_status Controller failover status. (0=No, 1=Yes)
# TYPE me5_controller_failover_status gauge
me5_controller_failover_status{id="A",ip="192.0.2.1"} 0
me5_controller_failover_status{id="B",ip="192.0.2.2"} 0
# HELP me5_controller_health Controller health status. (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A)
# TYPE me5_controller_health gauge
me5_controller_health{id="A",ip="192.0.2.1"} 0
me5_controller_health{id="B",ip="192.0.2.2"} 0
# HELP me5_controller_info Controller metadata.
# TYPE me5_controller_info gauge
me5_controller_info{cache_memory_size="16384",cpld_version="1.10",description="ASSY,CRD,CTL,TYPEB,12G,SAS,4",drive_bus_type="SAS",drive_channels="4",host_ports="4",hw_version="2.2",id="A",ip="192.0.2.1",mac="00:00:5E:00:53:01",mfg_date="2023-12-29 19:40:00",model="ME5084",part_number="55PKT",position="Left",redundancy_mode="Active-Active ULP",revision="A00",serial="CN0BBBBBBBBB0000BA00",system_memory_size="24576",vendor="DELL EMC",wwn="500C0FF000000000"} 1
me5_controller_info{cache_memory_size="16384",cpld_version="1.10",description="ASSY,CRD,CTL,TYPEB,12G,SAS,4",drive_bus_type="SAS",drive_channels="4",host_ports="4",hw_version="2.2",id="B",ip="192.0.2.2",mac="00:00:5E:00:53:02",mfg_date="2023-12-29 19:20:00",model="ME5084",part_number="55PKT",position="Right",redundancy_mode="Active-Active ULP",revision="A00",serial="CN0BBBBBBBBB0000DA00",system_memory_size="24576",vendor="DELL EMC",wwn="500C0FF000000000"} 1
# HELP me5_controller_redundancy_status Controller redundancy status. (0=Operational - Not Redundant, 2=Redundant, 4=Down, 5=Unknown)
# TYPE me5_controller_redundancy_status gauge
me5_controller_redundancy_status{id="A",ip="192.0.2.1"} 2
me5_controller_redundancy_status{id="B",ip="192.0.2.2"} 2
# HELP me5_controller_status Controller status. (0=Operational, 1=Down, 2=Not Installed)
# TYPE me5_controller_status gauge
me5_controller_status{id="A",ip="192.0.2.1"} 0
me5_controller_status{id="B",ip="192.0.2.2"} 0
# HELP me5_controller_storage_pools Number of virtual pools in the storage system.
# TYPE me5_controller_storage_pools gauge
me5_controller_storage_pools{id="A",ip="192.0.2.1"} 1
me5_controller_storage_pools{id="B",ip="192.0.2.2"} 1
# HELP me5_controller_virtual_disks Number of disk groups in the storage system.
# TYPE me5_controller_virtual_disks gauge
me5_controller_virtual_disks{id="A",ip="192.0.2.1"} 1
me5_controller_virtual_disks{id="B",ip="192.0.2.2"} 1
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_controller_disks",
		"me5_controller_failover_status",
		"me5_controller_health",
		"me5_controller_info",
		"me5_controller_redundancy_status",
		"me5_controller_status",
		"me5_controller_storage_pools",
		"me5_controller_virtual_disks",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectControllers_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/controllers": errors.New("connection refused")},
	}, onlyEnabled(CollectorControllers))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectControllers(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
