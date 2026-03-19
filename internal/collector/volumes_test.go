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

func TestCollectVolumes(t *testing.T) {
	data, err := os.ReadFile("testdata/volumes.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/volumes": data},
	}, onlyEnabled(CollectorVolumes))

	expected := `
# HELP me5_volume_allocated_bytes The amount of space currently allocated to a virtual volume, or the total size of a linear volume, in bytes.
# TYPE me5_volume_allocated_bytes gauge
me5_volume_allocated_bytes{name="projects",serial="00c0fffa39f700005f95066801000000"} 1.0007907139584e+14
me5_volume_allocated_bytes{name="scratch",serial="00c0fffa39f700006095066801000000"} 4.00820273152e+11
# HELP me5_volume_health Volume health status. (0=OK, 1=Degraded, 2=Fault, 3=Unknown)
# TYPE me5_volume_health gauge
me5_volume_health{name="projects",serial="00c0fffa39f700005f95066801000000"} 0
me5_volume_health{name="scratch",serial="00c0fffa39f700006095066801000000"} 0
# HELP me5_volume_info Volume information.
# TYPE me5_volume_info gauge
me5_volume_info{allowed_storage_tiers="Performance,Standard,Archive",creation_date_time="2025-04-21 18:58:39",description="",is_snapshot="No",name="projects",owner="A",parent_volume="",raidtype="N/A",serial="00c0fffa39f700005f95066801000000",storage_pool_name="A",storage_type="Virtual",tier_affinity="No Affinity",usage_type="Unspecified",virtual_disk_name="A",volume_class="standard",volume_group="UNGROUPEDVOLUMES",volume_type="base",wwn="600C0FF000FA39F75F95066801000000"} 1
me5_volume_info{allowed_storage_tiers="Performance,Standard,Archive",creation_date_time="2025-04-21 18:58:40",description="",is_snapshot="No",name="scratch",owner="A",parent_volume="",raidtype="N/A",serial="00c0fffa39f700006095066801000000",storage_pool_name="A",storage_type="Virtual",tier_affinity="No Affinity",usage_type="Unspecified",virtual_disk_name="A",volume_class="standard",volume_group="UNGROUPEDVOLUMES",volume_type="base",wwn="600C0FF000FA39F76095066801000000"} 1
# HELP me5_volume_metadata_bytes Amount of pool metadata currently being used by the volume, in bytes.
# TYPE me5_volume_metadata_bytes gauge
me5_volume_metadata_bytes{name="projects",serial="00c0fffa39f700005f95066801000000"} 1.90971904e+08
me5_volume_metadata_bytes{name="scratch",serial="00c0fffa39f700006095066801000000"} 4.030464e+06
# HELP me5_volume_size_bytes Total volume capacity, in bytes.
# TYPE me5_volume_size_bytes gauge
me5_volume_size_bytes{name="projects",serial="00c0fffa39f700005f95066801000000"} 1.00099996778496e+14
me5_volume_size_bytes{name="scratch",serial="00c0fffa39f700006095066801000000"} 3.9999999311872e+13
# HELP me5_volume_total_bytes The total size of the volume, in bytes.
# TYPE me5_volume_total_bytes gauge
me5_volume_total_bytes{name="projects",serial="00c0fffa39f700005f95066801000000"} 1.00099996778496e+14
me5_volume_total_bytes{name="scratch",serial="00c0fffa39f700006095066801000000"} 3.9999999311872e+13
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "me5_volume_allocated_bytes", "me5_volume_health", "me5_volume_info", "me5_volume_metadata_bytes", "me5_volume_size_bytes", "me5_volume_total_bytes"); err != nil {
		t.Error(err)
	}
}

func TestCollectVolumes_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/volumes": errors.New("connection refused")},
	}, onlyEnabled(CollectorVolumes))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectVolumes(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
