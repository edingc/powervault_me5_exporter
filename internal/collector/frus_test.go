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

func TestCollectFRUs(t *testing.T) {
	data, err := os.ReadFile("testdata/frus.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	c := NewME5Collector(&mockClient{
		responses: map[string][]byte{"/show/frus": data},
	}, onlyEnabled(CollectorFRUs))

	expected := `
# HELP me5_fru_info FRU metadata.
# TYPE me5_fru_info gauge
me5_fru_info{description="",enclosure_id="0",location="FAN MODULE SLOT 0",mfg_date="N/A",name="FAN MODULE",part_number="0998648-03",revision="",serial="SAWUXXXXXXXXXXXXX0"} 1
me5_fru_info{description="",enclosure_id="0",location="FAN MODULE SLOT 1",mfg_date="N/A",name="FAN MODULE",part_number="0998648-03",revision="",serial="SAWUXXXXXXXXXXXXX1"} 1
me5_fru_info{description="",enclosure_id="0",location="FAN MODULE SLOT 2",mfg_date="N/A",name="FAN MODULE",part_number="0998648-03",revision="",serial="SAWUXXXXXXXXXXXXX2"} 1
me5_fru_info{description="",enclosure_id="0",location="FAN MODULE SLOT 3",mfg_date="N/A",name="FAN MODULE",part_number="0998648-03",revision="",serial="SAWUXXXXXXXXXXXXX3"} 1
me5_fru_info{description="",enclosure_id="0",location="FAN MODULE SLOT 4",mfg_date="N/A",name="FAN MODULE",part_number="0998648-03",revision="",serial="SAWUXXXXXXXXXXXXX4"} 1
me5_fru_info{description="",enclosure_id="0",location="LOWER DRAWER LEFT SIDE",mfg_date="2023-02-07 14:09:00",name="SIDEPLANE",part_number="1014681-10",revision="",serial="BPSXXXXXXXXXXXXXX2"} 1
me5_fru_info{description="",enclosure_id="0",location="LOWER DRAWER RIGHT SIDE",mfg_date="2022-12-12 08:31:00",name="SIDEPLANE",part_number="1014687-08",revision="",serial="BPSXXXXXXXXXXXXXX3"} 1
me5_fru_info{description="",enclosure_id="0",location="UPPER DRAWER LEFT SIDE",mfg_date="2023-02-07 13:13:00",name="SIDEPLANE",part_number="1014681-10",revision="",serial="BPSXXXXXXXXXXXXXX0"} 1
me5_fru_info{description="",enclosure_id="0",location="UPPER DRAWER RIGHT SIDE",mfg_date="2022-12-13 11:55:00",name="SIDEPLANE",part_number="1014687-08",revision="",serial="BPSXXXXXXXXXXXXXX1"} 1
me5_fru_info{description="ASSY,CHAS,RKMNT,5U84NPCM",enclosure_id="0",location="MID-PLANE SLOT",mfg_date="2023-05-19 21:56:00",name="CHASSIS_MIDPLANE",part_number="J8GKY",revision="A01",serial="CN0AAAAAAAA000000AA01"} 1
me5_fru_info{description="ASSY,CRD,CTL,TYPEB,12G,SAS,4",enclosure_id="0",location="LEFT IOM SLOT",mfg_date="2023-12-29 19:40:00",name="RAID_IOM",part_number="55PKT",revision="A00",serial="CN0BBBBBBBBB0000BA00"} 1
me5_fru_info{description="ASSY,CRD,CTL,TYPEB,12G,SAS,4",enclosure_id="0",location="RIGHT IOM SLOT",mfg_date="2023-12-29 19:20:00",name="RAID_IOM",part_number="55PKT",revision="A00",serial="CN0BBBBBBBBB0000DA00"} 1
me5_fru_info{description="PWR SPLY,5U,ME4",enclosure_id="0",location="LEFT PSU SLOT",mfg_date="N/A",name="POWER_SUPPLY",part_number="6JN28",revision="A00",serial="CN0CCCCCCCC0000CC0001"} 1
me5_fru_info{description="PWR SPLY,5U,ME4",enclosure_id="0",location="RIGHT PSU SLOT",mfg_date="N/A",name="POWER_SUPPLY",part_number="6JN28",revision="A00",serial="CN0CCCCCCCC0000CC0002"} 1
# HELP me5_fru_status FRU status (0=Invalid Data, 1=Fault, 2= Absent, 3=Power Off, 4=OK).
# TYPE me5_fru_status gauge
me5_fru_status{name="CHASSIS_MIDPLANE",serial="CN0AAAAAAAA000000AA01"} 4
me5_fru_status{name="FAN MODULE",serial="SAWUXXXXXXXXXXXXX0"} 4
me5_fru_status{name="FAN MODULE",serial="SAWUXXXXXXXXXXXXX1"} 4
me5_fru_status{name="FAN MODULE",serial="SAWUXXXXXXXXXXXXX2"} 4
me5_fru_status{name="FAN MODULE",serial="SAWUXXXXXXXXXXXXX3"} 4
me5_fru_status{name="FAN MODULE",serial="SAWUXXXXXXXXXXXXX4"} 4
me5_fru_status{name="POWER_SUPPLY",serial="CN0CCCCCCCC0000CC0001"} 4
me5_fru_status{name="POWER_SUPPLY",serial="CN0CCCCCCCC0000CC0002"} 4
me5_fru_status{name="RAID_IOM",serial="CN0BBBBBBBBB0000BA00"} 4
me5_fru_status{name="RAID_IOM",serial="CN0BBBBBBBBB0000DA00"} 4
me5_fru_status{name="SIDEPLANE",serial="BPSXXXXXXXXXXXXXX0"} 4
me5_fru_status{name="SIDEPLANE",serial="BPSXXXXXXXXXXXXXX1"} 4
me5_fru_status{name="SIDEPLANE",serial="BPSXXXXXXXXXXXXXX2"} 4
me5_fru_status{name="SIDEPLANE",serial="BPSXXXXXXXXXXXXXX3"} 4
`

	if err := testutil.CollectAndCompare(c, strings.NewReader(expected),
		"me5_fru_info",
		"me5_fru_status",
	); err != nil {
		t.Error(err)
	}
}

func TestCollectFRUs_APIError(t *testing.T) {
	c := NewME5Collector(&mockClient{
		errs: map[string]error{"/show/frus": errors.New("connection refused")},
	}, onlyEnabled(CollectorFRUs))

	ch := make(chan prometheus.Metric, 10)
	if err := c.CollectFRUs(context.Background(), ch); err == nil {
		t.Error("expected error, got nil")
	}
	if len(ch) != 0 {
		t.Errorf("expected no metrics on error, got %d", len(ch))
	}
}
