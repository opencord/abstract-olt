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

package physical

import (
	"fmt"
)

/*
PONPort represents a single PON port on the OLT chassis
*/
type PONPort struct {
	Number   int
	DeviceID string
	Onts     [64]Ont
	Parent   *SimpleOLT `json:"-" bson:"-"`
}

/*
AllReadyActiveError - thrown when an attempt to activate a ONT which is already activated
*/
type AllReadyActiveError struct {
	slotNum    int
	clli       string
	ponportNum int
	ontNumber  int
}

/*
Error - the interface method that must be implemented on error
*/
func (e *AllReadyActiveError) Error() string {
	return fmt.Sprintf("Attempt to Activate ONT %d on PONPort %d Slot %d on %s but already active", e.ontNumber, e.ponportNum, e.slotNum, e.clli)
}

/*
AllReadyDeactivatedError - thrown when an attempt to activate a ONT which is already activated
*/
type AllReadyDeactivatedError struct {
	slotNum    int
	clli       string
	ponportNum int
	ontNumber  int
}

/*
Error - the interface method that must be implemented on error
*/
func (e *AllReadyDeactivatedError) Error() string {
	return fmt.Sprintf("Attempt to De-Activate ONT %d on PONPort %d Slot %d on %s but not active", e.ontNumber, e.ponportNum, e.slotNum, e.clli)
}

/*
ActivateOnt - passes ont information to chassis to make call to NEM to activate (whitelist) ont
*/
func (port *PONPort) ActivateOnt(number int, sVlan int, cVlan int, serialNumber string, nasPortID string, circuitID string) error {
	slot := port.Parent
	chassis := slot.Parent

	if port.Onts[number-1].Active {
		e := AllReadyActiveError{ontNumber: number, slotNum: slot.Number, ponportNum: port.Number, clli: chassis.CLLI}
		return &e
	}
	ont := Ont{Number: number, Svlan: sVlan, Cvlan: cVlan, SerialNumber: serialNumber, Parent: port, NasPortID: nasPortID, CircuitID: circuitID}
	port.Onts[number-1] = ont
	port.Parent.Parent.provisionONT(ont)
	port.Onts[number-1].Active = true
	return nil

}

/*
DeleteOnt - passes ont information to chassis to make call to NEM to de-activate (de-whitelist) ont
*/
func (port *PONPort) DeleteOnt(number int, sVlan int, cVlan int, serialNumber string) error {
	slot := port.Parent
	chassis := slot.Parent
	if port.Onts[number-1].Active != true {
		e := AllReadyDeactivatedError{ontNumber: number, slotNum: slot.Number, ponportNum: port.Number, clli: chassis.CLLI}
		return &e
	}
	ont := Ont{Number: number, Svlan: sVlan, Cvlan: cVlan, SerialNumber: serialNumber, Parent: port}
	chassis.deleteONT(ont)
	port.Onts[number-1].Active = false

	return nil
}
