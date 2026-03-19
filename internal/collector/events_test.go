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

func TestCollectEvents(t *testing.T) {
	data, err := os.ReadFile("testdata/events.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/events": data},
	}, onlyEnabled(CollectorEvents))

	expected := `
# HELP me5_events_by_severity Number of events grouped by severity.
# TYPE me5_events_by_severity gauge
me5_events_by_severity{severity="CRITICAL"} 1
me5_events_by_severity{severity="ERROR"} 1
me5_events_by_severity{severity="INFORMATIONAL"} 6
me5_events_by_severity{severity="RESOLVED"} 1
me5_events_by_severity{severity="WARNING"} 1
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "me5_events_by_severity"); err != nil {
		t.Error(err)
	}
}

func TestCollectEvents_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/events": errors.New("connection refused")},
	}, onlyEnabled(CollectorEvents))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectEvents(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
