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
	"strings"
	"testing"

	"gerrit.opencord.org/abstract-olt/models/tosca"
)

var expectedOutput = `tosca_definitions_version: tosca_simple_yaml_1_0
imports:
- custom_types/rcordsubscriber.yaml
description: Pre-provsion a subscriber
topology_template:
  node_templates:
    myName:
      type: tosca.nodes.RCORDSubscriber
      properties:
        name: myName
        status: pre-provisioned
        c_tag: 20
        s_tag: 2
        onu_device: onuSerialNumber
        nas_port_id: /1/1/1/1/1.9
        circuit_id: /1/1/1/1/1.9-CID
        remote_id: myCilli
`

var sub tosca.SubscriberProvision

func TestAddSubscriber_NewSubscriberProvision(t *testing.T) {
	sub = tosca.NewSubscriberProvision("myName", 20, 2, "onuSerialNumber", "/1/1/1/1/1.9", "/1/1/1/1/1.9-CID", "myCilli")
}

func TestAddSubscriber_ToYaml(t *testing.T) {
	y, err := sub.ToYaml()
	if err != nil {
		t.Fatalf("olt.ToYaml() failed with %v\n", err)
	}

	x := strings.Compare(y, expectedOutput)
	if x != 0 {
		t.Fatal("ToYaml didn't produce the expected yaml")
	}

}
