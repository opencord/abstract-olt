/*
 Copyright 2017 the original author or authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package abstract_test

import (
	"testing"

	"gerrit.opencord.org/abstract-olt/models/abstract"
)

func TestChassisUtils_GenerateChassis(t *testing.T) {
	chassis := abstract.GenerateChassis("MY_CLLI", 1, 1)
	slot := chassis.Slots[6]
	port := slot.Ports[0]
	ont := port.Onts[3]
	svlan := ont.Svlan
	cvlan := ont.Cvlan
	if svlan != 98 { // see map doc
		t.Errorf("SVlan should be 98 and is %d\n", svlan)
	}
	if cvlan != 434 { // see map doc
		t.Errorf("CVlan should be 434 and is %d\n", cvlan)
	}
}
