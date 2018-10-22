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
	PhysPort *physical.PONPort `json:"-"`
	Parent   *Slot             `json:"-"`
}

/*
UnprovisonedPortError - thrown when an attempt is made to address a physical port that hasn't been mapped to an abstract port
*/
type UnprovisonedPortError struct {
	oltNum  int
	clli    string
	portNum int
}

/*
Error - the interface method that must be implemented on error
*/
func (e *UnprovisonedPortError) Error() string {
	return fmt.Sprintf("Port %d for olt %d on AbstractChasis  %s is not provisioned", e.portNum, e.oltNum, e.clli)
}
func (port *Port) provisionOnt(ontNumber int, serialNumber string) error {

	slot := port.Parent
	chassis := slot.Parent
	baseID := fmt.Sprintf("%d/%d/%d/%d:%d.1.1", chassis.Rack, chassis.Shelf, slot.Number, port.Number, ontNumber)
	nasPortID := fmt.Sprintf("PON %s", baseID)
	circuitID := fmt.Sprintf("%s %s", chassis.CLLI, baseID)

	if port.PhysPort == nil {
		err := UnprovisonedPortError{oltNum: slot.Number, clli: chassis.CLLI, portNum: port.Number}
		return &err
	}
	phyPort := port.PhysPort
	ont := port.Onts[ontNumber-1]
	err := phyPort.ActivateOnt(ontNumber, ont.Svlan, ont.Cvlan, serialNumber, nasPortID, circuitID)
	return err
}
func (port *Port) provisionOntFull(ontNumber int, serialNumber string, cTag uint32, sTag uint32, nasPortID string, circuitID string) error {
	slot := port.Parent

	if port.PhysPort == nil {
		chassis := slot.Parent
		err := UnprovisonedPortError{oltNum: slot.Number, clli: chassis.CLLI, portNum: port.Number}
		return &err
	}
	phyPort := port.PhysPort
	err := phyPort.ActivateOnt(ontNumber, sTag, cTag, serialNumber, nasPortID, circuitID)
	return err
}
func (port *Port) deleteOnt(ontNumber int, serialNumber string) error {
	if port.PhysPort == nil {
		slot := port.Parent
		chassis := slot.Parent
		err := UnprovisonedPortError{oltNum: slot.Number, clli: chassis.CLLI, portNum: port.Number}
		return &err
	}
	phyPort := port.PhysPort
	ont := port.Onts[ontNumber-1]
	err := phyPort.DeleteOnt(ontNumber, ont.Svlan, ont.Cvlan, serialNumber)
	return err
}
