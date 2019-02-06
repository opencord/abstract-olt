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
package inventory

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"gerrit.opencord.org/abstract-olt/models"
	"gerrit.opencord.org/abstract-olt/models/physical"
)

type Chassis struct {
	Clli      string
	Rack      int
	Shelf     int
	XOSAddr   net.TCPAddr
	LineCards []LineCard
}
type LineCard struct {
	Number int
	Olts   []PhysicalOlt
}
type PhysicalOlt struct {
	Address  net.TCPAddr
	Hostname string
	Ports    []Port
}
type Port struct {
	AbstractNumber int
	PhysicalNumber int
	Onts           []Ont
}
type Ont struct {
	Number       int
	Active       bool
	SVlan        uint32
	CVlan        uint32
	SerialNumber string
	NasPortID    string
	CircuitID    string
}

func GatherAllInventory() string {
	chassisMap := models.GetChassisMap()
	chassis_s := []Chassis{}
	for clli, chassisHolder := range *chassisMap {
		chassis := parseClli(clli, chassisHolder)
		chassis_s = append(chassis_s, chassis)
	}
	bytes, _ := json.Marshal(chassis_s)
	return string(bytes)
}

func GatherInventory(clli string) (string, error) {
	if clli == "" {
		return "", errors.New("You must provide a CLLI")
	}
	chassisMap := models.GetChassisMap()
	chassisHolder := (*chassisMap)[clli]
	if chassisHolder == nil {
		errorMsg := fmt.Sprintf("No Chassis Holder found for CLLI %s", clli)
		return "", errors.New(errorMsg)
	}
	chassis := parseClli(clli, chassisHolder)
	bytes, _ := json.Marshal(chassis)
	return string(bytes), nil
}

func parseClli(clli string, chassisHolder *models.ChassisHolder) Chassis {
	abstract := chassisHolder.AbstractChassis
	chassis := Chassis{}
	chassis.Clli = clli
	chassis.Rack = abstract.Rack
	chassis.Shelf = abstract.Shelf
	chassis.XOSAddr = chassisHolder.PhysicalChassis.XOSAddress

	lineCards := []LineCard{}
	for index, slot := range abstract.Slots {
		if slot.Ports[0].PhysPort != nil {
			lineCard := LineCard{Number: index + 1}
			var currentOLT *physical.SimpleOLT
			var physicalOLT PhysicalOlt
			var ports []Port
			olts := []PhysicalOlt{}
			for i := 0; i < 16; i++ {
				ponPort := slot.Ports[i].PhysPort
				if ponPort != nil {
					parentOLT := ponPort.Parent
					if currentOLT != parentOLT {
						if currentOLT != nil {
							physicalOLT.Ports = ports
							olts = append(olts, physicalOLT)
						}
						physicalOLT = PhysicalOlt{Address: parentOLT.Address, Hostname: parentOLT.Hostname}
						currentOLT = parentOLT
						ports = []Port{}

					}
					port := Port{AbstractNumber: i + 1, PhysicalNumber: ponPort.Number}
					onts := []Ont{}
					for _, physicalONT := range ponPort.Onts {
						if physicalONT.CircuitID != "" {
							ont := Ont{Number: physicalONT.Number, Active: physicalONT.Active, SVlan: physicalONT.Svlan, CVlan: physicalONT.Cvlan, SerialNumber: physicalONT.SerialNumber,
								NasPortID: physicalONT.NasPortID, CircuitID: physicalONT.CircuitID}
							onts = append(onts, ont)
						}
					}
					port.Onts = onts
					ports = append(ports, port)
				}

				if i == 15 { // last one
					physicalOLT.Ports = ports
					olts = append(olts, physicalOLT)
				}
			}
			lineCard.Olts = olts
			lineCards = append(lineCards, lineCard)
		}
	}
	chassis.LineCards = lineCards
	return chassis
}
