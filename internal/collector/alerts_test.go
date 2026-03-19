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

// "Collect" alerts from fixtures and ensure we get expected results
func TestCollectAlerts(t *testing.T) {
	data, err := os.ReadFile("testdata/alerts.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/alerts": data},
	}, onlyEnabled(CollectorAlerts))

	expected := `
# HELP me5_alerts_by_severity Number of alerts by severity, resolved and acknowledged state.
# TYPE me5_alerts_by_severity gauge
me5_alerts_by_severity{acknowledged="No",resolved="No",severity="INFORMATIONAL"} 3
me5_alerts_by_severity{acknowledged="No",resolved="Yes",severity="INFORMATIONAL"} 4
me5_alerts_by_severity{acknowledged="No",resolved="Yes",severity="WARNING"} 4
me5_alerts_by_severity{acknowledged="Yes",resolved="No",severity="INFORMATIONAL"} 6
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "me5_alerts_by_severity"); err != nil {
		t.Error(err)
	}
}

// Test what happens if API connection is faulty
func TestCollectAlerts_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/alerts": errors.New("connection refused")},
	}, onlyEnabled(CollectorAlerts))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectAlerts(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
