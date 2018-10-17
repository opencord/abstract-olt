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

package models

import (
	"encoding/json"
	"log"

	"gerrit.opencord.org/abstract-olt/internal/pkg/settings"
	"gerrit.opencord.org/abstract-olt/models/abstract"
	"gerrit.opencord.org/abstract-olt/models/physical"
)

func (chassisHolder ChassisHolder) Serialize() ([]byte, error) {
	return json.Marshal(chassisHolder.PhysicalChassis)

}

func (chassisHolder *ChassisHolder) Deserialize(jsonData []byte) error {
	physicalChassis := physical.Chassis{}
	err := json.Unmarshal(jsonData, &physicalChassis)
	if err != nil {
		return err
	}
	abstractChassis := abstract.GenerateChassis(physicalChassis.CLLI, 1, 1)
	chassisHolder.AbstractChassis = abstractChassis
	chassisHolder.PhysicalChassis = physicalChassis

	//first handle abstract parent pointers
	for i := 0; i < len(abstractChassis.Slots); i++ {
		slot := &abstractChassis.Slots[i]
		slot.Parent = &abstractChassis
		for j := 0; j < len(slot.Ports); j++ {
			port := &slot.Ports[j]
			port.Parent = slot
			for k := 0; k < len(port.Onts); k++ {
				port.Onts[k].Parent = port
			}
		}
	}
	//second handle physical parent pointers
	for i := 0; i < len(physicalChassis.Linecards); i++ {
		slot := physicalChassis.Linecards[i]
		slot.Parent = &physicalChassis
		for j := 0; j < len(slot.Ports); j++ {
			port := &slot.Ports[j]
			port.Parent = &slot
			for k := 0; k < len(port.Onts); k++ {
				port.Onts[k].Parent = port
			}
		}
	}
	//finally handle abstract.Port -> physical.PonPort pointers

	for i := 0; i < len(physicalChassis.Linecards); i++ {
		slot := physicalChassis.Linecards[i]
		for j := 0; j < len(slot.Ports); j++ {
			absPort, _ := chassisHolder.AbstractChassis.NextPort()
			absPort.PhysPort = &slot.Ports[j]
		}
	}
	if settings.GetDebug() {
		log.Printf("created chassis %v\n", abstractChassis)
	}
	return nil
}
