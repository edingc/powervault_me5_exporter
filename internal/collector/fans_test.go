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

func TestCollectFans(t *testing.T) {
	data, err := os.ReadFile("testdata/fans.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/fans": data},
	}, onlyEnabled(CollectorFans))

	expected := `
# HELP me5_fan_health Fan health status. (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A)
# TYPE me5_fan_health gauge
me5_fan_health{location="Enclosure 0, Fan Module 0",name="Fan 0"} 0
me5_fan_health{location="Enclosure 0, Fan Module 0",name="Fan 1"} 0
# HELP me5_fan_speed Fan speed (revolutions per minute).
# TYPE me5_fan_speed gauge
me5_fan_speed{location="Enclosure 0, Fan Module 0",name="Fan 0"} 7440
me5_fan_speed{location="Enclosure 0, Fan Module 0",name="Fan 1"} 7440
# HELP me5_fan_status Fan unit status. (0=Up, 1=Error, 2=Off, 3=Missing)
# TYPE me5_fan_status gauge
me5_fan_status{location="Enclosure 0, Fan Module 0",name="Fan 0"} 0
me5_fan_status{location="Enclosure 0, Fan Module 0",name="Fan 1"} 0
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "me5_fan_health", "me5_fan_speed", "me5_fan_status"); err != nil {
		t.Error(err)
	}
}

func TestCollectFans_EmptyResponse(t *testing.T) {
	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/fans": []byte(`{"fan":[]}`)},
	}, onlyEnabled(CollectorFans))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectFans(context.Background(), ch); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics for empty response, got %d", len(ch))
	}
}

func TestCollectFans_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/fans": errors.New("connection refused")},
	}, onlyEnabled(CollectorFans))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectFans(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
