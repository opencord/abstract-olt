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
	"testing"

	"gerrit.opencord.org/abstract-olt/models/tosca"
)

var deleteExpected = `tosca_definitions_version: tosca_simple_yaml_1_0
imports:
- custom_types/onudevice.yaml
description: Delete an ont
topology_template:
  node_templates:
    device#onu:
      type: tosca.nodes.ONUDevice
      properties:
        serial_number: some_serial
`
var ontToDelete tosca.OntDelete

//func NewOntProvision(serialNumber string, oltIP net.IP, ponPortNumber int) OntProvision {

func TestOntDelete_NewOntDelete(t *testing.T) {
	ontToDelete = tosca.NewOntDelete("some_serial")
	ontYaml, _ := ontToDelete.ToYaml()
	if ontYaml != deleteExpected {
		t.Fatalf("Didn't generate the expected yaml\n Generated:\n%s \nExpected:\n%s\n", ontYaml, deleteExpected)
	}
}
