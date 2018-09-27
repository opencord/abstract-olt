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
	"encoding/json"
)

func (chassis *Chassis) Serialize() ([]byte, error) {
	return json.Marshal(chassis)
}

func (chassis *Chassis) Deserialize(jsonData []byte) error {
	err := json.Unmarshal(jsonData, chassis)

	for i := 0; i < len(chassis.Slots); i++ {
		var slot *Slot
		slot = &chassis.Slots[i]
		slot.Parent = chassis
		for j := 0; j < len(slot.Ports); j++ {
			var port *Port
			port = &slot.Ports[j]
			port.Parent = slot
			for k := 0; k < len(port.Onts); k++ {
				port.Onts[k].Parent = port
			}
		}
	}

	return err
}
