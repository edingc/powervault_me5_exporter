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

func TestCollectFirmwareBundles(t *testing.T) {
	data, err := os.ReadFile("testdata/firmware-bundles.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/firmware-bundles": data},
	}, onlyEnabled(CollectorFirmwareBundles))

	expected := `
# HELP me5_firmware_bundle_health Firmware bundle health (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).
# TYPE me5_firmware_bundle_health gauge
me5_firmware_bundle_health{version="ME5.1.2.1.1"} 0
me5_firmware_bundle_health{version="ME5.1.2.1.5"} 0
# HELP me5_firmware_bundle_info Firmware bundle metadata.
# TYPE me5_firmware_bundle_info gauge
me5_firmware_bundle_info{build_date="2024-10-02T18:07:19Z",status_name="Available",version="ME5.1.2.1.1"} 1
me5_firmware_bundle_info{build_date="2025-07-21T07:30:34Z",status_name="Active",version="ME5.1.2.1.5"} 1
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_firmware_bundle_health",
		"me5_firmware_bundle_info",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectFirmwareBundles_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/firmware-bundles": errors.New("connection refused")},
	}, onlyEnabled(CollectorFirmwareBundles))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectFirmwareBundles(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
