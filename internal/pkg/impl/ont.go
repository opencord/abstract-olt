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

package impl

import (
	"errors"
	"fmt"

	"gerrit.opencord.org/abstract-olt/models"
)

/*
ProvisionOnt - provisions ont using sTag,cTag,NasPortID, and CircuitID generated internally
*/
func ProvisionOnt(clli string, slotNumber int, portNumber int, ontNumber int, serialNumber string) (bool, error) {
	myChan := getSyncChannel()
	<-myChan
	defer done(myChan, true)
	chassisMap := models.GetChassisMap()
	chassisHolder := (*chassisMap)[clli]
	if chassisHolder == nil {
		errString := fmt.Sprintf("There is no chassis with CLLI of %s", clli)
		return false, errors.New(errString)
	}
	err := chassisHolder.AbstractChassis.ActivateONT(slotNumber, portNumber, ontNumber, serialNumber)
	isDirty = true
	return true, err
}

/*
ActivateSerial - provisions ont using sTag,cTag,NasPortID, and CircuitID generated internally
*/
func ActivateSerial(clli string, slotNumber int, portNumber int, ontNumber int, serialNumber string) (bool, error) {
	myChan := getSyncChannel()
	<-myChan
	defer done(myChan, true)
	chassisMap := models.GetChassisMap()
	chassisHolder := (*chassisMap)[clli]
	if chassisHolder == nil {
		errString := fmt.Sprintf("There is no chassis with CLLI of %s", clli)
		return false, errors.New(errString)
	}
	err := chassisHolder.AbstractChassis.ActivateSerial(slotNumber, portNumber, ontNumber, serialNumber)
	isDirty = true
	return true, err
}

/*
PreProvisionOnt - provisions ont using sTag,cTag,NasPortID, and CircuitID passed in
*/
func PreProvisionOnt(clli string, slotNumber int, portNumber int, ontNumber int, cTag uint32, sTag uint32, nasPortID string, circuitID string, techProfile string, speedProfile string) (bool, error) {
	myChan := getSyncChannel()
	<-myChan
	defer done(myChan, true)
	chassisMap := models.GetChassisMap()
	chassisHolder := (*chassisMap)[clli]
	if chassisHolder == nil {
		errString := fmt.Sprintf("There is no chassis with CLLI of %s", clli)
		return false, errors.New(errString)
	}
	err := chassisHolder.AbstractChassis.PreProvisonONT(slotNumber, portNumber, ontNumber, cTag, sTag, nasPortID, circuitID, techProfile, speedProfile)
	isDirty = true
	return true, err
}

/*
ProvisionOntFull - provisions ont using sTag,cTag,NasPortID, and CircuitID passed in
*/
func ProvisionOntFull(clli string, slotNumber int, portNumber int, ontNumber int, serialNumber string, cTag uint32, sTag uint32, nasPortID string, circuitID string) (bool, error) {
	myChan := getSyncChannel()
	<-myChan
	defer done(myChan, true)
	chassisMap := models.GetChassisMap()
	chassisHolder := (*chassisMap)[clli]
	if chassisHolder == nil {
		errString := fmt.Sprintf("There is no chassis with CLLI of %s", clli)
		return false, errors.New(errString)
	}
	err := chassisHolder.AbstractChassis.ActivateONTFull(slotNumber, portNumber, ontNumber, serialNumber, cTag, sTag, nasPortID, circuitID)
	isDirty = true
	return true, err
}

/*
DeleteOnt - deletes a previously provision ont
*/
func DeleteOnt(clli string, slotNumber int, portNumber int, ontNumber int, serialNumber string) (bool, error) {
	myChan := getSyncChannel()
	<-myChan
	defer done(myChan, true)
	chassisMap := models.GetChassisMap()
	chassisHolder := (*chassisMap)[clli]
	err := chassisHolder.AbstractChassis.DeleteONT(slotNumber, portNumber, ontNumber, serialNumber)
	isDirty = true
	return true, err
}
