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
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	"gerrit.opencord.org/abstract-olt/internal/pkg/settings"
	"gerrit.opencord.org/abstract-olt/models"
)

/*
ChangeXOSUserPassword - allows update of xos credentials
*/
func ChangeXOSUserPassword(clli string, xosUser string, xosPassword string) (bool, error) {
	myChan := getSyncChannel()
	<-myChan
	defer done(myChan, true)
	chassisMap := models.GetChassisMap()
	chassisHolder := (*chassisMap)[clli]
	if chassisHolder == nil {
		errString := fmt.Sprintf("There is no chassis with CLLI of %s", clli)
		return false, errors.New(errString)
	}
	xosIP := chassisHolder.PhysicalChassis.XOSAddress.IP
	xosPort := chassisHolder.PhysicalChassis.XOSAddress.Port
	loginWorked := testLogin(xosUser, xosPassword, xosIP, xosPort)
	if !loginWorked {
		return false, errors.New("Unable to validate login when changing password")
	}

	chassisHolder.PhysicalChassis.XOSUser = xosUser
	chassisHolder.PhysicalChassis.XOSPassword = xosPassword
	isDirty = true
	return true, nil

}

func testLogin(xosUser string, xosPassword string, xosIP net.IP, xosPort int) bool {
	if settings.GetDummy() {
		return true
	}
	if settings.GetGrpc() {
		return true
	}
	var dummyYaml = `
tosca_definitions_version: tosca_simple_yaml_1_0
imports:
  - custom_types/site.yaml
description: anything
topology_template:
  node_templates:
    mysite:
      type: tosca.nodes.Site
      properties:
        must-exist: true
        name: mysite
`
	client := &http.Client{}
	requestList := fmt.Sprintf("http://%s:%d/run", xosIP, xosPort)
	req, err := http.NewRequest("POST", requestList, strings.NewReader(dummyYaml))
	req.Header.Add("xos-username", xosUser)
	req.Header.Add("xos-password", xosPassword)
	resp, err := client.Do(req)
	log.Printf("testLogin resp:%v", resp)
	if err != nil {
		log.Printf("Unable to validate XOS Login Information %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		fmt.Println(bodyString)
		return true
	}
	return false

}
