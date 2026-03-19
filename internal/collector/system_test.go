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

func TestCollectSystem(t *testing.T) {
	data, err := os.ReadFile("testdata/system.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/system": data},
	}, onlyEnabled(CollectorSystem))

	expected := `
# HELP me5_system_controller_a_status System controller A status (0=Operational, 1=Down, 2=Not Installed).
# TYPE me5_system_controller_a_status gauge
me5_system_controller_a_status{serial="CN0BBBBBBBBB0000BA00"} 0
# HELP me5_system_controller_b_status System controller B status (0=Operational, 1=Down, 2=Not Installed).
# TYPE me5_system_controller_b_status gauge
me5_system_controller_b_status{serial="CN0BBBBBBBBB0000DA00"} 0
# HELP me5_system_enclosure_count The number of enclosures in the system.
# TYPE me5_system_enclosure_count gauge
me5_system_enclosure_count 1
# HELP me5_system_fde_security_status Full Disk Encryption security status (1=Unsecured, 2=Secured, 3=Secured - Lock Ready, 4=Secure - Locked).
# TYPE me5_system_fde_security_status gauge
me5_system_fde_security_status 1
# HELP me5_system_health Overall system health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).
# TYPE me5_system_health gauge
me5_system_health 3
# HELP me5_system_info System metadata.
# TYPE me5_system_info gauge
me5_system_info{midplane_serial="CN0AAAAAAAA000000AA01",product_brand="PowerVault",product_id="ME5084",system_contact="Uninitialized Contact",system_information="Uninitialized Info",system_location="Uninitialized Location",system_name="Uninitialized Name",vendor_name="DELL EMC"} 1
# HELP me5_system_redundancy_mode System redundancy mode (8=Active-Active ULP, 9=Single Controller, 10=Failed Over, 11=Down).
# TYPE me5_system_redundancy_mode gauge
me5_system_redundancy_mode 8
# HELP me5_system_redundancy_status System redundancy status (0=Operational but not redundant, 2=Redundant, 4=Down, 5=Unknown).
# TYPE me5_system_redundancy_status gauge
me5_system_redundancy_status 2
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_system_controller_a_status",
		"me5_system_controller_b_status",
		"me5_system_enclosure_count",
		"me5_system_fde_security_status",
		"me5_system_health",
		"me5_system_info",
		"me5_system_redundancy_mode",
		"me5_system_redundancy_status",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectSystem_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/system": errors.New("connection refused")},
	}, onlyEnabled(CollectorSystem))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectSystem(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
