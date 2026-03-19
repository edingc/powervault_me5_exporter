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

func TestCollectSessions(t *testing.T) {
	data, err := os.ReadFile("testdata/sessions.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/sessions": data},
	}, onlyEnabled(CollectorSessions))

	expected := `
# HELP me5_sessions_active Number of active sessions on the storage system.
# TYPE me5_sessions_active gauge
me5_sessions_active 1
# HELP me5_sessions_info Active session metadata.
# TYPE me5_sessions_info gauge
me5_sessions_info{host="192.168.1.2",interface="wbi",session_id="93",username="monitoring"} 1
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "me5_sessions_active", "me5_sessions_info"); err != nil {
		t.Error(err)
	}
}

func TestCollectSessions_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/sessions": errors.New("connection refused")},
	}, onlyEnabled(CollectorSessions))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectSessions(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
