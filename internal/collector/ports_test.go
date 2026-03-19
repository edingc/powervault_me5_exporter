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

func TestCollectPorts(t *testing.T) {
	data, err := os.ReadFile("testdata/ports.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/ports": data},
	}, onlyEnabled(CollectorPorts))

	expected := `
# HELP me5_port_actual_speed_numeric Actual link speed (0=1Gb, 1=2Gb, 2=4Gb, 3=Port Disconnected, 6=6Gb, 7=8Gb, 8=10Mb, 9=100Mb, 11=12Gb, 12=16Gb).
# TYPE me5_port_actual_speed_numeric gauge
me5_port_actual_speed_numeric{controller="A",port="A0"} 3
me5_port_actual_speed_numeric{controller="A",port="A1"} 3
me5_port_actual_speed_numeric{controller="A",port="A2"} 3
me5_port_actual_speed_numeric{controller="A",port="A3"} 6
me5_port_actual_speed_numeric{controller="B",port="B0"} 3
me5_port_actual_speed_numeric{controller="B",port="B1"} 3
me5_port_actual_speed_numeric{controller="B",port="B2"} 3
me5_port_actual_speed_numeric{controller="B",port="B3"} 6
# HELP me5_port_health Port health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).
# TYPE me5_port_health gauge
me5_port_health{controller="A",port="A0"} 0
me5_port_health{controller="A",port="A1"} 0
me5_port_health{controller="A",port="A2"} 0
me5_port_health{controller="A",port="A3"} 0
me5_port_health{controller="B",port="B0"} 0
me5_port_health{controller="B",port="B1"} 0
me5_port_health{controller="B",port="B2"} 0
me5_port_health{controller="B",port="B3"} 0
# HELP me5_port_info Port metadata.
# TYPE me5_port_info gauge
me5_port_info{configured_speed="",controller="A",media="SAS",port="A0",target_id="500c0ff000000000",type="SAS"} 1
me5_port_info{configured_speed="",controller="A",media="SAS",port="A1",target_id="500c0ff000000100",type="SAS"} 1
me5_port_info{configured_speed="",controller="A",media="SAS",port="A2",target_id="500c0ff000000200",type="SAS"} 1
me5_port_info{configured_speed="",controller="A",media="SAS",port="A3",target_id="500c0ff000000300",type="SAS"} 1
me5_port_info{configured_speed="",controller="B",media="SAS",port="B0",target_id="500c0ff000000400",type="SAS"} 1
me5_port_info{configured_speed="",controller="B",media="SAS",port="B1",target_id="500c0ff000000500",type="SAS"} 1
me5_port_info{configured_speed="",controller="B",media="SAS",port="B2",target_id="500c0ff000000600",type="SAS"} 1
me5_port_info{configured_speed="",controller="B",media="SAS",port="B3",target_id="500c0ff000000700",type="SAS"} 1
# HELP me5_port_sas_active_lanes Number of active SAS lanes.
# TYPE me5_port_sas_active_lanes gauge
me5_port_sas_active_lanes{controller="A",port="A0"} 0
me5_port_sas_active_lanes{controller="A",port="A1"} 0
me5_port_sas_active_lanes{controller="A",port="A2"} 0
me5_port_sas_active_lanes{controller="A",port="A3"} 4
me5_port_sas_active_lanes{controller="B",port="B0"} 0
me5_port_sas_active_lanes{controller="B",port="B1"} 0
me5_port_sas_active_lanes{controller="B",port="B2"} 0
me5_port_sas_active_lanes{controller="B",port="B3"} 4
# HELP me5_port_sas_expected_lanes Number of expected SAS lanes.
# TYPE me5_port_sas_expected_lanes gauge
me5_port_sas_expected_lanes{controller="A",port="A0"} 4
me5_port_sas_expected_lanes{controller="A",port="A1"} 4
me5_port_sas_expected_lanes{controller="A",port="A2"} 4
me5_port_sas_expected_lanes{controller="A",port="A3"} 4
me5_port_sas_expected_lanes{controller="B",port="B0"} 4
me5_port_sas_expected_lanes{controller="B",port="B1"} 4
me5_port_sas_expected_lanes{controller="B",port="B2"} 4
me5_port_sas_expected_lanes{controller="B",port="B3"} 4
# HELP me5_port_status Port status (0=Up, 1=Warning, 2=Error, 3=Not Present, 6=Disconnected).
# TYPE me5_port_status gauge
me5_port_status{controller="A",port="A0"} 6
me5_port_status{controller="A",port="A1"} 6
me5_port_status{controller="A",port="A2"} 6
me5_port_status{controller="A",port="A3"} 0
me5_port_status{controller="B",port="B0"} 6
me5_port_status{controller="B",port="B1"} 6
me5_port_status{controller="B",port="B2"} 6
me5_port_status{controller="B",port="B3"} 0
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_port_actual_speed_numeric",
		"me5_port_health",
		"me5_port_info",
		"me5_port_sas_active_lanes",
		"me5_port_sas_expected_lanes",
		"me5_port_status",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectPorts_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/ports": errors.New("connection refused")},
	}, onlyEnabled(CollectorPorts))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectPorts(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
