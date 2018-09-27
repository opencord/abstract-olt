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

import (
	"errors"
)

const MAX_SLOTS int = 16
const MAX_PORTS int = 16

/*
Chassis is a model that takes up to 16 discreet OLT chassis as if it is a 16 slot OLT chassis
*/
type Chassis struct {
	CLLI      string
	Slots     [16]Slot
	Rack      int
	Shelf     int
	AllocInfo PortAllocationInfo
}

type PortAllocationInfo struct {
	// Current info on next port to be allocated
	slot       int
	port       int
	outOfPorts bool
}

func (chassis *Chassis) NextPort() (*Port, error) {
	info := &chassis.AllocInfo

	if info.outOfPorts {
		return nil, errors.New("Abstract chassis out of ports")
	}

	nextPort := &chassis.Slots[info.slot].Ports[info.port]

	info.port++
	if info.port == MAX_PORTS {
		info.port = 0
		info.slot++
		if info.slot == MAX_SLOTS {
			info.slot = 0
			info.outOfPorts = true
		}
	}

	return nextPort, nil
}
func (chassis *Chassis) ActivateONT(slotNumber int, portNumber int, ontNumber int, serialNumber string) error {
	err := chassis.Slots[slotNumber-1].Ports[portNumber-1].provisionOnt(ontNumber, serialNumber)
	return err
}
func (chassis *Chassis) DeleteONT(slotNumber int, portNumber int, ontNumber int, serialNumber string) error {
	err := chassis.Slots[slotNumber-1].Ports[portNumber-1].deleteOnt(ontNumber, serialNumber)
	return err
}
