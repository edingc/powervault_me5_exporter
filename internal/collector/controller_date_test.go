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

func TestCollectControllerDate(t *testing.T) {
	data, err := os.ReadFile("testdata/controller-date.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/controller-date": data},
	}, onlyEnabled(CollectorControllerDate))

	expected := `
# HELP me5_controller_date_date_time_numeric Controller date and time as a Unix timestamp.
# TYPE me5_controller_date_date_time_numeric gauge
me5_controller_date_date_time_numeric{ntp_address="192.168.1.100",time_zone_region="GMT"} 1.773674743e+09
# HELP me5_controller_date_ntp_contact_time_info NTP metadata.
# TYPE me5_controller_date_ntp_contact_time_info gauge
me5_controller_date_ntp_contact_time_info{ntp_address="192.168.1.100",ntp_last_contact_time="2026-03-13 17:54:25",ntp_state="Enabled"} 1
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "me5_controller_date_date_time_numeric", "me5_controller_date_ntp_contact_time_info"); err != nil {
		t.Error(err)
	}
}

func TestCollectControllerDate_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/controller-date": errors.New("connection refused")},
	}, onlyEnabled(CollectorControllerDate))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectControllerDate(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
