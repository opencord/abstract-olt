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

package abstract

/*
GenerateChassis - constructs a new AbstractOLT Chassis
*/
func GenerateChassis(CLLI string) *Chassis {
	chassis := Chassis{CLLI: CLLI}

	var slots [16]Slot
	for i := 0; i < 16; i++ {
		slots[i] = generateSlot(i, &chassis)
	}

	chassis.Slots = slots
	return &chassis
}

func generateSlot(n int, c *Chassis) Slot {
	slot := Slot{Number: n, Parent: c}

	var ports [16]Port
	for i := 0; i < 16; i++ {
		ports[i] = generatePort(i, &slot)
	}

	slot.Ports = ports
	return slot
}
func generatePort(n int, s *Slot) Port {
	port := Port{Number: n, Parent: s}

	var onts [64]Ont
	for i := 0; i < 64; i++ {
		/* adding one because the system that provisions is 1 based on everything not 0 based*/
		onts[i] = Ont{Number: i, Svlan: calculateSvlan(s.Number+1, n+1, i+1),
			Cvlan: calculateCvlan(s.Number+1, n+1, i+1), Parent: &port}
	}

	port.Onts = onts
	return port
}

func calculateCvlan(slot int, port int, ont int) int {
	ontPortOffset := 120 // Max(ONT_SLOT) * Max(ONT_PORT) = 10 * 12 = 120
	ontSlotOffset := 12  //= Max(ONT_PORT) = 12
	vlanOffset := 1      //(VID 1 is reserved)

	cVid := ((ont-1)%32)*ontPortOffset +
		(slot-1)*ontSlotOffset + port + vlanOffset

	return cVid
}

func calculateSvlan(slot int, port int, ont int) int {
	ltSlotOffset := 16
	vlanGap := 288  // Max(LT_SLOT) * Max(ltSlotOffset) = 18 * 16 = 288
	vlanOffset := 1 //(VID 1 is reserved)

	sVid := ((slot-1)*ltSlotOffset + port) + ((ont-1)/32)*vlanGap + vlanOffset

	return sVid
}

/*
NextPort pulls the first unMapped port in the abstract chassis so the next physical port can be mapped to it
*/
