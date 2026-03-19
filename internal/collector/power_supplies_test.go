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

func TestCollectPowerSupplies(t *testing.T) {
	data, err := os.ReadFile("testdata/power-supplies.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/power-supplies": data},
	}, onlyEnabled(CollectorPowerSupplies))

	expected := `
# HELP me5_power_supply_health Power supply health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).
# TYPE me5_power_supply_health gauge
me5_power_supply_health{name="PSU 0, Left",serial="CN0CCCCCCCC0000CC0001"} 0
me5_power_supply_health{name="PSU 1, Right",serial="CN0CCCCCCCC0000CC0002"} 0
# HELP me5_power_supply_info Power supply metadata.
# TYPE me5_power_supply_info gauge
me5_power_supply_info{description="PWR SPLY,5U,ME4",fw_revision="0306",location="Enclosure 0 - Left",model="6JN28",name="PSU 0, Left",part_number="6JN28",serial="CN0CCCCCCCC0000CC0001"} 1
me5_power_supply_info{description="PWR SPLY,5U,ME4",fw_revision="0306",location="Enclosure 0 - Right",model="6JN28",name="PSU 1, Right",part_number="6JN28",serial="CN0CCCCCCCC0000CC0002"} 1
# HELP me5_power_supply_status Power supply status (0=Up, 1=Warning, 2=Error, 3=Not Present, 4=Unknown).
# TYPE me5_power_supply_status gauge
me5_power_supply_status{name="PSU 0, Left",serial="CN0CCCCCCCC0000CC0001"} 0
me5_power_supply_status{name="PSU 1, Right",serial="CN0CCCCCCCC0000CC0002"} 0
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "me5_power_supply_health", "me5_power_supply_info", "me5_power_supply_status"); err != nil {
		t.Error(err)
	}
}

func TestCollectPowerSupplies_EmptyResponse(t *testing.T) {
	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/power-supplies": []byte(`{"power-supplies":[]}`)},
	}, onlyEnabled(CollectorPowerSupplies))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectPowerSupplies(context.Background(), ch); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics for empty response, got %d", len(ch))
	}
}

func TestCollectPowerSupplies_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/power-supplies": errors.New("connection refused")},
	}, onlyEnabled(CollectorPowerSupplies))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectPowerSupplies(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
