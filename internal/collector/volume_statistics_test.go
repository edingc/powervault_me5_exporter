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

func TestCollectVolumeStats(t *testing.T) {
	data, err := os.ReadFile("testdata/volume-statistics.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/volume-statistics": data},
	}, onlyEnabled(CollectorVolumeStats))

	expected := `
# HELP me5_volume_allocated_pages The number of pages allocated to the volume.
# TYPE me5_volume_allocated_pages gauge
me5_volume_allocated_pages{name="projects",serial="00c0fffa39f700005f95066801000000"} 2.386071e+07
me5_volume_allocated_pages{name="scratch",serial="00c0fffa39f700006095066801000000"} 95563
# HELP me5_volume_bytes_per_second The data transfer rate, in bytes per second, calculated over the interval since these statistics were last requested or reset.
# TYPE me5_volume_bytes_per_second gauge
me5_volume_bytes_per_second{name="projects",serial="00c0fffa39f700005f95066801000000"} 1.677375488e+09
me5_volume_bytes_per_second{name="scratch",serial="00c0fffa39f700006095066801000000"} 0
# HELP me5_volume_data_read_bytes_total The amount of data read since these statistics were last reset or since the controller was restarted.
# TYPE me5_volume_data_read_bytes_total counter
me5_volume_data_read_bytes_total{name="projects",serial="00c0fffa39f700005f95066801000000"} 1.21642827222784e+15
me5_volume_data_read_bytes_total{name="scratch",serial="00c0fffa39f700006095066801000000"} 5.4784210944e+10
# HELP me5_volume_data_written_bytes_total The amount of data written since these statistics were last reset or since the controller was restarted.
# TYPE me5_volume_data_written_bytes_total counter
me5_volume_data_written_bytes_total{name="projects",serial="00c0fffa39f700005f95066801000000"} 2.124249812672512e+15
me5_volume_data_written_bytes_total{name="scratch",serial="00c0fffa39f700006095066801000000"} 2.31084371968e+11
# HELP me5_volume_iops Input/output operations per second, calculated over the interval since these statistics were last requested or reset.
# TYPE me5_volume_iops gauge
me5_volume_iops{name="projects",serial="00c0fffa39f700005f95066801000000"} 18412
me5_volume_iops{name="scratch",serial="00c0fffa39f700006095066801000000"} 0
# HELP me5_volume_read_cache_hits_total For the controller that owns the volume, the number of times the block to be read is found in cache.
# TYPE me5_volume_read_cache_hits_total counter
me5_volume_read_cache_hits_total{name="projects",serial="00c0fffa39f700005f95066801000000"} 6.7637966885e+10
me5_volume_read_cache_hits_total{name="scratch",serial="00c0fffa39f700006095066801000000"} 2.700044e+06
# HELP me5_volume_read_cache_misses_total For the controller that owns the volume, the number of times the block to be read is not found in cache.
# TYPE me5_volume_read_cache_misses_total counter
me5_volume_read_cache_misses_total{name="projects",serial="00c0fffa39f700005f95066801000000"} 1.0134925632e+10
me5_volume_read_cache_misses_total{name="scratch",serial="00c0fffa39f700006095066801000000"} 2.947432e+06
# HELP me5_volume_reads_total The number of read operations since these statistics were last reset or since the controller was restarted.
# TYPE me5_volume_reads_total counter
me5_volume_reads_total{name="projects",serial="00c0fffa39f700005f95066801000000"} 4.724453694e+09
me5_volume_reads_total{name="scratch",serial="00c0fffa39f700006095066801000000"} 3.152118e+06
# HELP me5_volume_write_cache_hits_total For the controller that owns the volume, the number of times the block written to is found in cache.
# TYPE me5_volume_write_cache_hits_total counter
me5_volume_write_cache_hits_total{name="projects",serial="00c0fffa39f700005f95066801000000"} 7.7642825913e+10
me5_volume_write_cache_hits_total{name="scratch",serial="00c0fffa39f700006095066801000000"} 4.0293053e+07
# HELP me5_volume_write_cache_misses_total For the controller that owns the volume, the number of times the block written to is not found in cache.
# TYPE me5_volume_write_cache_misses_total counter
me5_volume_write_cache_misses_total{name="projects",serial="00c0fffa39f700005f95066801000000"} 1.29569276577e+11
me5_volume_write_cache_misses_total{name="scratch",serial="00c0fffa39f700006095066801000000"} 9.283882e+06
# HELP me5_volume_writes_total The number of write operations since these statistics were last reset or since the controller was restarted.
# TYPE me5_volume_writes_total counter
me5_volume_writes_total{name="projects",serial="00c0fffa39f700005f95066801000000"} 3.2872668285e+10
me5_volume_writes_total{name="scratch",serial="00c0fffa39f700006095066801000000"} 4.4624264e+07
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_volume_allocated_pages",
		"me5_volume_bytes_per_second",
		"me5_volume_data_read_bytes_total",
		"me5_volume_data_written_bytes_total",
		"me5_volume_iops",
		"me5_volume_read_cache_hits_total",
		"me5_volume_read_cache_misses_total",
		"me5_volume_reads_total",
		"me5_volume_write_cache_hits_total",
		"me5_volume_write_cache_misses_total",
		"me5_volume_writes_total",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectVolumeStats_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/volume-statistics": errors.New("connection refused")},
	}, onlyEnabled(CollectorVolumeStats))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectVolumeStats(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
