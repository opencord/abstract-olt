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

package integration

import (
	"testing"

	"gerrit.opencord.org/abstract-olt/internal/pkg/chassisSerialize"
	"gerrit.opencord.org/abstract-olt/models/abstract"
)

func TestSerialize(t *testing.T) {
	chassis1 := generateTestChassis()
	bytes1, err1 := chassisSerialize.Serialize(chassis1)
	chassis2, err2 := chassisSerialize.Deserialize(bytes1)
	bytes2, err3 := chassisSerialize.Serialize(chassis2)
	chassis3, err4 := chassisSerialize.Deserialize(bytes2)

	ok(t, err1)
	ok(t, err2)
	ok(t, err3)
	ok(t, err4)
	equals(t, chassis1, chassis3)
	equals(t, chassis3.Slots[2].Parent, chassis3)
	equals(t, chassis3.Slots[15].Ports[8].Parent, &chassis3.Slots[15])
	equals(t, chassis3.Slots[0].Ports[10].Onts[15].Parent, &chassis3.Slots[0].Ports[10])
}

func generateTestChassis() *abstract.Chassis {
	chassis := abstract.GenerateChassis("My_CLLI", 1, 1)

	var slots [16]abstract.Slot
	for i := 0; i < 16; i++ {
		slots[i] = generateTestSlot(i, chassis)
	}

	chassis.Slots = slots
	return chassis
}

func generateTestSlot(n int, c *abstract.Chassis) abstract.Slot {
	slot := abstract.Slot{Number: n, Parent: c}

	var ports [16]abstract.Port
	for i := 0; i < 16; i++ {
		ports[i] = generateTestPort(16*n+i, &slot)
	}

	slot.Ports = ports
	return slot
}

func generateTestPort(n int, s *abstract.Slot) abstract.Port {
	port := abstract.Port{Number: n, Parent: s}

	var onts [64]abstract.Ont
	for i := 0; i < 64; i++ {
		j := n*64 + i
		onts[i] = abstract.Ont{Number: j, Svlan: j * 10, Cvlan: j*10 + 5, Parent: &port}
	}

	port.Onts = onts
	return port
}
