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

import "gerrit.opencord.org/abstract-olt/models"

/*
Reflow - takes internal config and resends to xos
*/
func Reflow() (bool, error) {
	myChan := getSyncChannel()
	<-myChan
	defer done(myChan, true)
	chassisMap := models.GetChassisMap()
	for _, chassisHolder := range *chassisMap {
		physical := chassisHolder.PhysicalChassis
		for index := range physical.Linecards {
			olt := physical.Linecards[index]
			physical.SendOltTosca(olt)
			for portIndex := range olt.Ports {
				port := olt.Ports[portIndex]
				for ontIndex := range port.Onts {
					ont := port.Onts[ontIndex]
					if ont.Active {
						physical.SendOntTosca(ont)
						physical.SendSubscriberTosca(ont)

					}

				}

			}
		}
	}
	return true, nil
	//TODO lots of this could throw errors
}
