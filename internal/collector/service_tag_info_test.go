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

func TestCollectServiceTagInfo(t *testing.T) {
	data, err := os.ReadFile("testdata/service-tag-info.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/service-tag-info": data},
	}, onlyEnabled(CollectorServiceTag))

	expected := `
# HELP me5_service_tag_info Enclosure metadata.
# TYPE me5_service_tag_info gauge
me5_service_tag_info{enclosure_id="0",service_tag="AAAAAAA"} 1
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "me5_service_tag_info"); err != nil {
		t.Error(err)
	}
}

func TestCollectServiceTagInfo_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/service-tag-info": errors.New("connection refused")},
	}, onlyEnabled(CollectorServiceTag))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectServiceTag(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
