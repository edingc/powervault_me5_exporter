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

func TestCollectSnapshotSpace(t *testing.T) {
	data, err := os.ReadFile("testdata/snapshot-space.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/snapshot-space": data},
	}, onlyEnabled(CollectorSnapshotSpace))

	// allocated-size-numeric: 2678784 * 512 = 1371537408
	// snap-limit-size-numeric: 34631528448 * 512 = 17731342565376
	expected := `
# HELP me5_snapshot_space_allocated_bytes Snapshot space currently allocated in bytes.
# TYPE me5_snapshot_space_allocated_bytes gauge
me5_snapshot_space_allocated_bytes{pool="A",serial="00c0fffa39f700001395066801000000"} 1.371537408e+09
# HELP me5_snapshot_space_limit_bytes Snapshot space limit in bytes.
# TYPE me5_snapshot_space_limit_bytes gauge
me5_snapshot_space_limit_bytes{pool="A",serial="00c0fffa39f700001395066801000000"} 1.7731342565376e+13
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "me5_snapshot_space_allocated_bytes", "me5_snapshot_space_limit_bytes"); err != nil {
		t.Error(err)
	}
}

func TestCollectSnapshotSpace_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/snapshot-space": errors.New("connection refused")},
	}, onlyEnabled(CollectorSnapshotSpace))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectSnapshotSpace(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
