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
	"net"
	"testing"

	"gerrit.opencord.org/abstract-olt/models/tosca"
)

var expected = `tosca_definitions_version: tosca_simple_yaml_1_0
imports:
- custom_types/attworkflowdriverwhitelistentry.yaml
- custom_types/attworkflowdriverservice.yaml
description: Create an entry in the whitelist
topology_template:
  node_templates:
    service#att:
      type: tosca.nodes.AttWorkflowDriverService
      properties:
        name: att-workflow-driver
        must-exist: true
    some_serial:
      type: tosca.nodes.AttWorkflowDriverWhiteListEntry
      properties:
        serial_number: some_serial
        pon_port_id: 536870913
        device_id: of:00000000c0a8010b
      requirements:
      - owner:
          node: service#att
          relationship: tosca.relationships.BelongsToOne
`
var ont tosca.OntProvision

//func NewOntProvision(serialNumber string, oltIP net.IP, ponPortNumber int) OntProvision {

func TestOntProvision_NewOntProvision(t *testing.T) {
	ont = tosca.NewOntProvision("some_serial", net.ParseIP("192.168.1.11"), 2)
	ontYaml, _ := ont.ToYaml()
	if ontYaml != expected {
		t.Fatalf("Didn't generate the expected yaml\n Generated:\n%s \nExpected:\n%s\n", ontYaml, expected)
	}
}
