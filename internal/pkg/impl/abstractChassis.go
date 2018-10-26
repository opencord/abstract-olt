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
	"log"
	"net"
	"strings"

	"gerrit.opencord.org/abstract-olt/internal/pkg/settings"
	"gerrit.opencord.org/abstract-olt/models"
	"gerrit.opencord.org/abstract-olt/models/abstract"
	"gerrit.opencord.org/abstract-olt/models/physical"
)

/*
CreateChassis - allocates a new Chassis struct and stores it in chassisMap
*/
func CreateChassis(clli string, xosAddress net.TCPAddr, xosUser string, xosPassword string, shelf int, rack int) (string, error) {
	myChan := getSyncChannel()
	<-myChan
	defer done(myChan, true)
	chassisMap := models.GetChassisMap()

	loginWorked := testLogin(xosUser, xosPassword, xosAddress.IP, xosAddress.Port)
	if !loginWorked {
		return "", errors.New("Unable to validate login not creating Abstract Chassis")
	}

	chassisHolder := (*chassisMap)[clli]
	if chassisHolder != nil {
		errMsg := fmt.Sprintf("AbstractChassis %s already exists", clli)
		return "", errors.New(errMsg)
	}

	abstractChassis := abstract.GenerateChassis(clli, rack, shelf)
	phyChassis := physical.Chassis{CLLI: clli, XOSUser: xosUser, XOSPassword: xosPassword, XOSAddress: xosAddress, Rack: rack, Shelf: shelf}

	chassisHolder = &models.ChassisHolder{AbstractChassis: abstractChassis, PhysicalChassis: phyChassis}
	if settings.GetDebug() {
		output := fmt.Sprintf("%v", abstractChassis)
		formatted := strings.Replace(output, "{", "\n{", -1)
		log.Printf("new chassis %s\n", formatted)
	}
	(*chassisMap)[clli] = chassisHolder
	isDirty = true
	return clli, nil
}
