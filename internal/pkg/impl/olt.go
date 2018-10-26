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
	"net"

	"gerrit.opencord.org/abstract-olt/models"
	"gerrit.opencord.org/abstract-olt/models/physical"
)

/*
CreateOLTChassis adds an OLT chassis/line card to the Physical chassis
*/
func CreateOLTChassis(clli string, oltType string, address net.TCPAddr, hostname string) (string, error) {
	myChan := getSyncChannel()
	<-myChan
	defer done(myChan, true)
	chassisMap := models.GetChassisMap()
	chassisHolder := (*chassisMap)[clli]
	if chassisHolder == nil {
		errString := fmt.Sprintf("There is no chassis with CLLI of %s", clli)
		return "", errors.New(errString)
	}
	physicalChassis := &chassisHolder.PhysicalChassis
	sOlt := physical.SimpleOLT{CLLI: clli, Hostname: hostname, Address: address, Parent: physicalChassis}
	switch oltType {
	case "edgecore":
		sOlt.CreateEdgecore()
	case "adtran":
	case "tibit":
	}
	physicalChassis.AddOLTChassis(sOlt)
	ports := sOlt.GetPorts()
	for i := 0; i < len(ports); i++ {
		absPort, _ := chassisHolder.AbstractChassis.NextPort()

		absPort.PhysPort = &ports[i]
		//AssignTraits(&ports[i], absPort)
	}
	isDirty = true
	return clli, nil

}
