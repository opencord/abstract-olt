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
	"fmt"

	"gerrit.opencord.org/abstract-olt/models/physical"
)

/*
Port represents a single PON port on the OLT chassis
*/
type Port struct {
	Number int
	// DeviceID string
	Onts     [64]Ont
	PhysPort *physical.PONPort
	Parent   *Slot `json:"-"`
}

type UnprovisonedPortError struct {
	oltNum  int
	clli    string
	portNum int
}

func (e *UnprovisonedPortError) Error() string {
	return fmt.Sprintf("Port %d for olt %d on AbstractChasis  %s is not provisioned", e.portNum, e.oltNum, e.clli)
}
func (port *Port) provisionOnt(ontNumber int, serialNumber string) error {
	if port.PhysPort == nil {
		slot := port.Parent
		chassis := slot.Parent
		err := UnprovisonedPortError{oltNum: slot.Number, clli: chassis.CLLI, portNum: port.Number}
		return &err
	}

	phyPort := port.PhysPort
	ont := port.Onts[ontNumber-1]
	phyPort.ActivateOnt(ontNumber, ont.Svlan, ont.Cvlan, serialNumber)
	return nil
}
