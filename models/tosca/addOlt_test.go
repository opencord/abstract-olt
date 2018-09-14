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
package tosca_test

import (
	"fmt"
	"strings"
	"testing"

	"gerrit.opencord.org/abstract-olt/models/tosca"
)

var output = `tosca_definitions_version: tosca_simple_yaml_1_0
imports:
- custom_types/oltdevice.yaml
- custom_types/onudevice.yaml
- custom_types/ponport.yaml
- custom_types/voltservice.yaml
description: Create a simulated OLT Device in VOLTHA
topology_template:
  node_templates:
    service#volt:
      type: tosca.nodes.VOLTService
      properties:
        name: volt
        must-exist: true
    olt_device:
      type: tosca.nodes.OLTDevice
      properties:
        name: myName
        device_type: openolt
        host: 192.168.1.1
        port: 9191
        outer_tpid: "0x8100"
        uplink: "65536"
        nas_id: my_clli
      requirements:
      - volt_service:
          node: service#volt
          relationship: tosca.relationships.BelongsToOne
`

var olt tosca.OltProvsion

func TestAddOlt_NewOltProvsion(t *testing.T) {
	fmt.Println("In TestAddOlt_NewOltProvsion")
	olt = tosca.NewOltProvision("my_clli", "myName", "openolt", "192.168.1.1", 9191)
	fmt.Printf("%v\n\n", olt)
}

func TestAddOlt_ToYaml(t *testing.T) {
	y, err := olt.ToYaml()
	if err != nil {
		t.Fatalf("olt.ToYaml() failed with %v\n", err)
	}
	x := strings.Compare(y, output)
	if x != 0 {
		t.Fatal("ToYaml didn't produce the expected yaml")
	}
	fmt.Printf("Compare is %d\n", x)

	fmt.Printf(y)
	fmt.Println("******")
	fmt.Print(output)
	fmt.Println("******")
}
