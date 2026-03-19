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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// mockClient is a test double for APIClient. Add paths to responses to return
// fixture data, or to errs to simulate API failures.
type mockClient struct {
	responses map[string][]byte
	errs      map[string]error
}

func (m *mockClient) Get(_ context.Context, path string, dest any) error {
	if err, ok := m.errs[path]; ok {
		return err
	}
	if data, ok := m.responses[path]; ok {
		return json.Unmarshal(data, dest)
	}
	return fmt.Errorf("unexpected path: %s", path)
}

// TestCollect_SubCollectorError verifies that Collect still emits scrape metrics when
// a sub-collector fails, and exercises the error branches in Collect and collect.
func TestCollect_SubCollectorError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/alerts": errors.New("connection refused")},
	}, map[string]bool{CollectorAlerts: true})

	ch := make(chan prometheus.Metric, 100)
	c.Collect(ch)
	close(ch)

	var count int
	for range ch {
		count++
	}
	// Collect always emits scrape_success and scrape_duration_seconds.
	if count == 0 {
		t.Error("expected scrape metrics to be emitted even on sub-collector error")
	}
}

// onlyEnabled returns an enabled map with only the named collectors active and all others disabled.
func onlyEnabled(names ...string) map[string]bool {
	m := make(map[string]bool, len(AllCollectors))
	for k := range AllCollectors {
		m[k] = false
	}
	for _, n := range names {
		m[n] = true
	}
	return m
}
