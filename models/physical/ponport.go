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
PreProvisionOnt - passes ont information to chassis to make call to NEM to activate (whitelist) ont
*/
func (port *PONPort) PreProvisionOnt(number int, sVlan uint32, cVlan uint32, nasPortID string, circuitID string, techProfile string, speedProfile string) error {
	fmt.Printf("PrPreProvisionOnt(number %d, sVlan %d, cVlan %d, nasPortID %s, circuitID %s, techProfile %s, speedProfile %s\n", number, sVlan, cVlan, nasPortID, circuitID, techProfile, speedProfile)
	slot := port.Parent
	chassis := slot.Parent

	if port.Onts[number-1].Active {
		e := AllReadyActiveError{ontNumber: number, slotNum: slot.Number, ponportNum: port.Number, clli: chassis.CLLI}
		return &e
	}
	ont := &port.Onts[number-1]
	ont.Number = number
	ont.Svlan = sVlan
	ont.Cvlan = cVlan
	ont.Parent = port
	ont.NasPortID = nasPortID
	ont.CircuitID = circuitID
	ont.TechProfile = techProfile
	ont.SpeedProfile = speedProfile
	fmt.Printf("ponPort PreProvision ont :%v\n", ont)
	return nil
}

/*
ActivateSerial - passes ont information to chassis to make call to NEM to activate (whitelist) ont assumes pre provisioned ont
*/
func (port *PONPort) ActivateSerial(number int, serialNumber string) error {
	slot := port.Parent
	chassis := slot.Parent

	if port.Onts[number-1].Active {
		e := AllReadyActiveError{ontNumber: number, slotNum: slot.Number, ponportNum: port.Number, clli: chassis.CLLI}
		return &e
	}
	ont := &port.Onts[number-1]
	ont.SerialNumber = serialNumber
	fmt.Println(ont)
	port.Parent.Parent.provisionONT(*ont)
	port.Onts[number-1].Active = true
	return nil

}

/*
ActivateOnt - passes ont information to chassis to make call to NEM to activate (whitelist) ont
*/
func (port *PONPort) ActivateOnt(number int, sVlan uint32, cVlan uint32, serialNumber string, nasPortID string, circuitID string) error {
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
func (port *PONPort) DeleteOnt(number int, sVlan uint32, cVlan uint32, serialNumber string) error {

	fmt.Printf("DeleteOnt(number %d, sVlan %d, cVlan %d, serialNumber %s)\n", number, sVlan, cVlan, serialNumber)
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
