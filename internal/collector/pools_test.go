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

func TestCollectPools(t *testing.T) {
	data, err := os.ReadFile("testdata/pools.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/pools": data},
	}, onlyEnabled(CollectorPools))

	expected := `
# HELP me5_pool_allocated_pages For a virtual pool, the number of 4 MB pages that are currently in use. For a linear pool, 0.
# TYPE me5_pool_allocated_pages gauge
me5_pool_allocated_pages{pool="A",serial="00c0fffa39f700001395066801000000"} 2.3956275e+07
# HELP me5_pool_available_bytes The available capacity in the pool, in bytes.
# TYPE me5_pool_available_bytes gauge
me5_pool_available_bytes{pool="A",serial="00c0fffa39f700001395066801000000"} 7.6833529790464e+13
# HELP me5_pool_available_pages For a virtual pool, the number of 4 MB pages that are still available to be allocated. For a linear pool, 0.
# TYPE me5_pool_available_pages gauge
me5_pool_available_pages{pool="A",serial="00c0fffa39f700001395066801000000"} 1.8318541e+07
# HELP me5_pool_health Pool health status (0=OK, 1=Degraded, 2=Fault, 3=Unknown, 4=N/A).
# TYPE me5_pool_health gauge
me5_pool_health{pool="A",serial="00c0fffa39f700001395066801000000"} 0
# HELP me5_pool_info Pool metadata.
# TYPE me5_pool_info gauge
me5_pool_info{owner="A",pool="A",preferred_owner="A",sector_format="512e",serial="00c0fffa39f700001395066801000000",storage_type="Virtual"} 1
# HELP me5_pool_overcommitted Whether the pool is overcommitted (0=No, 1=Yes).
# TYPE me5_pool_overcommitted gauge
me5_pool_overcommitted{pool="A",serial="00c0fffa39f700001395066801000000"} 0
# HELP me5_pool_rfc_size_bytes The total size of the read cache in the pool, in bytes.
# TYPE me5_pool_rfc_size_bytes gauge
me5_pool_rfc_size_bytes{pool="A",serial="00c0fffa39f700001395066801000000"} 0
# HELP me5_pool_size_bytes The total capacity of the pool, in bytes.
# TYPE me5_pool_size_bytes gauge
me5_pool_size_bytes{pool="A",serial="00c0fffa39f700001395066801000000"} 1.77313429848064e+14
# HELP me5_pool_volumes The number of volumes in the pool.
# TYPE me5_pool_volumes gauge
me5_pool_volumes{pool="A",serial="00c0fffa39f700001395066801000000"} 2
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_pool_allocated_pages",
		"me5_pool_available_bytes",
		"me5_pool_available_pages",
		"me5_pool_health",
		"me5_pool_info",
		"me5_pool_overcommitted",
		"me5_pool_rfc_size_bytes",
		"me5_pool_size_bytes",
		"me5_pool_volumes",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectPools_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/pools": errors.New("connection refused")},
	}, onlyEnabled(CollectorPools))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectPools(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
