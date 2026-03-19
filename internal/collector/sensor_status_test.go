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

func TestCollectSensors(t *testing.T) {
	data, err := os.ReadFile("testdata/sensor-status.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/sensor-status": data},
	}, onlyEnabled(CollectorSensors))

	expected := `
# HELP me5_sensor_status Sensor status (0=Unsupported, 1=OK, 2=Critical, 3=Warning, 4=Unrecoverable, 5=Not Installed, 6=Unknown, 7=Unavailable).
# TYPE me5_sensor_status gauge
me5_sensor_status{container="controllers",controller="A",id="sensor_temp_ctrl_A.5",name="Disk Controller Temperature-Ctlr A",type="Temperature"} 1
me5_sensor_status{container="controllers",controller="A",id="sensor_temp_ctrl_A.6",name="Host Controller Temperature-Ctlr A",type="Temperature"} 1
me5_sensor_status{container="controllers",controller="B",id="sensor_temp_ctrl_B.1",name="CPU Temperature-Ctlr B",type="Temperature"} 1
me5_sensor_status{container="controllers",controller="B",id="sensor_volt_ctrl_B.0",name="Capacitor Pack Voltage-Ctlr B",type="Voltage"} 1
me5_sensor_status{container="controllers",controller="B",id="sensor_volt_ctrl_B.1",name="Capacitor Cell 1 Voltage-Ctlr B",type="Voltage"} 1
me5_sensor_status{container="controllers",controller="B",id="sensor_volt_ctrl_B.2",name="Capacitor Cell 2 Voltage-Ctlr B",type="Voltage"} 1
me5_sensor_status{container="power-supplies",controller="N/A",id="sensor_curr_psu_0.0.0",name="Current 12V Rail Loc: left-PSU",type="Current"} 1
me5_sensor_status{container="power-supplies",controller="N/A",id="sensor_curr_psu_0.0.1",name="Current 5V Rail Loc: left-PSU",type="Current"} 1
# HELP me5_sensor_value Sensor value.
# TYPE me5_sensor_value gauge
me5_sensor_value{container="controllers",controller="A",id="sensor_temp_ctrl_A.5",name="Disk Controller Temperature-Ctlr A",type="Temperature"} 104
me5_sensor_value{container="controllers",controller="A",id="sensor_temp_ctrl_A.6",name="Host Controller Temperature-Ctlr A",type="Temperature"} 111
me5_sensor_value{container="controllers",controller="B",id="sensor_temp_ctrl_B.1",name="CPU Temperature-Ctlr B",type="Temperature"} 125
me5_sensor_value{container="controllers",controller="B",id="sensor_volt_ctrl_B.0",name="Capacitor Pack Voltage-Ctlr B",type="Voltage"} 10.75
me5_sensor_value{container="controllers",controller="B",id="sensor_volt_ctrl_B.1",name="Capacitor Cell 1 Voltage-Ctlr B",type="Voltage"} 2.69
me5_sensor_value{container="controllers",controller="B",id="sensor_volt_ctrl_B.2",name="Capacitor Cell 2 Voltage-Ctlr B",type="Voltage"} 2.69
me5_sensor_value{container="power-supplies",controller="N/A",id="sensor_curr_psu_0.0.0",name="Current 12V Rail Loc: left-PSU",type="Current"} 19.68
me5_sensor_value{container="power-supplies",controller="N/A",id="sensor_curr_psu_0.0.1",name="Current 5V Rail Loc: left-PSU",type="Current"} 0.07
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_sensor_status",
		"me5_sensor_value",
	); err != nil {
		t.Error(err)
	}
}

// TestCollectSensors_ParseFloatOverflow covers the else branch at sensor_status.go:73.
// The regex matches the digit string, but strconv.ParseFloat returns ErrRange because the
// value overflows float64, so only me5_sensor_status is emitted (not me5_sensor_value).
func TestCollectSensors_ParseFloatOverflow(t *testing.T) {
	// A 400-digit number overflows float64, triggering strconv.ParseFloat's ErrRange.
	overflow := `{"sensors":[{"durable-id":"sensor_temp_ctrl_A.0","controller-id":"A","sensor-name":"Test","value":"` +
		string(make([]byte, 400)) + `","status-numeric":1,"container":"controllers","sensor-type":"Temperature"}]}`
	// Replace null bytes with '9'
	b := []byte(overflow)
	for i := range b {
		if b[i] == 0 {
			b[i] = '9'
		}
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/sensor-status": b},
	}, onlyEnabled(CollectorSensors))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectSensors(context.Background(), ch); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Only me5_sensor_status emitted; me5_sensor_value skipped due to ParseFloat error.
	if len(ch) != 1 {
		t.Errorf("expected 1 metric (status only), got %d", len(ch))
	}
}

func TestCollectSensors_EmptyResponse(t *testing.T) {
	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/sensor-status": []byte(`{"sensors":[]}`)},
	}, onlyEnabled(CollectorSensors))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectSensors(context.Background(), ch); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics for empty response, got %d", len(ch))
	}
}

func TestCollectSensors_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/sensor-status": errors.New("connection refused")},
	}, onlyEnabled(CollectorSensors))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectSensors(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
