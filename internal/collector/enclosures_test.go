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

func TestCollectEnclosures(t *testing.T) {
	data, err := os.ReadFile("testdata/enclosures.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/enclosures": data},
	}, onlyEnabled(CollectorEnclosures))

	expected := `
# HELP me5_enclosure_cooling_element_count Number of fan units in the enclosure.
# TYPE me5_enclosure_cooling_element_count gauge
me5_enclosure_cooling_element_count{enclosure_id="0"} 10
# HELP me5_enclosure_disk_count Number of disk slots (not installed disks) in the enclosure.
# TYPE me5_enclosure_disk_count gauge
me5_enclosure_disk_count{enclosure_id="0"} 28
# HELP me5_enclosure_drawer_health Drawer health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).
# TYPE me5_enclosure_drawer_health gauge
me5_enclosure_drawer_health{drawer_id="0",enclosure_id="0"} 0
me5_enclosure_drawer_health{drawer_id="1",enclosure_id="0"} 0
# HELP me5_enclosure_drawer_status Drawer status (0=Unsupported, 1=OK, 2=Critical, 3=Warning, 4=Unrecoverable, 5=Not Installed, 6=Unknown, 7=Unavailable).
# TYPE me5_enclosure_drawer_status gauge
me5_enclosure_drawer_status{drawer_id="0",enclosure_id="0"} 1
me5_enclosure_drawer_status{drawer_id="1",enclosure_id="0"} 1
# HELP me5_enclosure_expander_health Expander health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).
# TYPE me5_enclosure_expander_health gauge
me5_enclosure_expander_health{drawer_id="0",enclosure_id="0",location="Enclosure 0, Drawer 0, Left Sideplane",name="Sideplane 24-port Expander 0"} 0
me5_enclosure_expander_health{drawer_id="0",enclosure_id="0",location="Enclosure 0, Drawer 0, Left Sideplane",name="Sideplane 36-port Expander 1"} 0
me5_enclosure_expander_health{drawer_id="0",enclosure_id="0",location="Enclosure 0, Drawer 0, Right Sideplane",name="Sideplane 24-port Expander 0"} 0
me5_enclosure_expander_health{drawer_id="0",enclosure_id="0",location="Enclosure 0, Drawer 0, Right Sideplane",name="Sideplane 36-port Expander 1"} 0
me5_enclosure_expander_health{drawer_id="1",enclosure_id="0",location="Enclosure 0, Drawer 1, Left Sideplane",name="Sideplane 24-port Expander 0"} 0
me5_enclosure_expander_health{drawer_id="1",enclosure_id="0",location="Enclosure 0, Drawer 1, Left Sideplane",name="Sideplane 36-port Expander 1"} 0
me5_enclosure_expander_health{drawer_id="1",enclosure_id="0",location="Enclosure 0, Drawer 1, Right Sideplane",name="Sideplane 24-port Expander 0"} 0
me5_enclosure_expander_health{drawer_id="1",enclosure_id="0",location="Enclosure 0, Drawer 1, Right Sideplane",name="Sideplane 36-port Expander 1"} 0
# HELP me5_enclosure_expander_info Expander metadata.
# TYPE me5_enclosure_expander_info gauge
me5_enclosure_expander_info{drawer_id="0",enclosure_id="0",location="Enclosure 0, Drawer 0, Left Sideplane",name="Sideplane 24-port Expander 0",revision="5375"} 1
me5_enclosure_expander_info{drawer_id="0",enclosure_id="0",location="Enclosure 0, Drawer 0, Left Sideplane",name="Sideplane 36-port Expander 1",revision="5375"} 1
me5_enclosure_expander_info{drawer_id="0",enclosure_id="0",location="Enclosure 0, Drawer 0, Right Sideplane",name="Sideplane 24-port Expander 0",revision="5375"} 1
me5_enclosure_expander_info{drawer_id="0",enclosure_id="0",location="Enclosure 0, Drawer 0, Right Sideplane",name="Sideplane 36-port Expander 1",revision="5375"} 1
me5_enclosure_expander_info{drawer_id="1",enclosure_id="0",location="Enclosure 0, Drawer 1, Left Sideplane",name="Sideplane 24-port Expander 0",revision="5375"} 1
me5_enclosure_expander_info{drawer_id="1",enclosure_id="0",location="Enclosure 0, Drawer 1, Left Sideplane",name="Sideplane 36-port Expander 1",revision="5375"} 1
me5_enclosure_expander_info{drawer_id="1",enclosure_id="0",location="Enclosure 0, Drawer 1, Right Sideplane",name="Sideplane 24-port Expander 0",revision="5375"} 1
me5_enclosure_expander_info{drawer_id="1",enclosure_id="0",location="Enclosure 0, Drawer 1, Right Sideplane",name="Sideplane 36-port Expander 1",revision="5375"} 1
# HELP me5_enclosure_health Enclosure health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).
# TYPE me5_enclosure_health gauge
me5_enclosure_health{enclosure_id="0"} 0
# HELP me5_enclosure_info Enclosure metadata.
# TYPE me5_enclosure_info gauge
me5_enclosure_info{enclosure_id="0",midplane_serial="CN0AAAAAAAA000000AA01",model="Array584_ME5SAS",revision="A01",vendor="DellEMC",wwn="500C0FF000000000"} 1
# HELP me5_enclosure_power_supply_count Number of power supplies in the enclosure.
# TYPE me5_enclosure_power_supply_count gauge
me5_enclosure_power_supply_count{enclosure_id="0"} 2
# HELP me5_enclosure_slot_count Number of disk slots in this enclosure.
# TYPE me5_enclosure_slot_count gauge
me5_enclosure_slot_count{enclosure_id="0"} 84
# HELP me5_enclosure_status Enclosure status (0=Unsupported, 1=OK, 2=Critical, 3=Warning, 4=Unrecoverable, 5=Not Installed, 6=Unknown, 7=Unavailable).
# TYPE me5_enclosure_status gauge
me5_enclosure_status{enclosure_id="0"} 1
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_enclosure_cooling_element_count",
		"me5_enclosure_disk_count",
		"me5_enclosure_drawer_health",
		"me5_enclosure_drawer_status",
		"me5_enclosure_expander_health",
		"me5_enclosure_expander_info",
		"me5_enclosure_health",
		"me5_enclosure_info",
		"me5_enclosure_power_supply_count",
		"me5_enclosure_slot_count",
		"me5_enclosure_status",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectEnclosures_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/enclosures": errors.New("connection refused")},
	}, onlyEnabled(CollectorEnclosures))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectEnclosures(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
