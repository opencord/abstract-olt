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
package tosca

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

// type: tosca.nodes.ONUDevice
var ontDeleteTemplate = `tosca_definitions_version: tosca_simple_yaml_1_0
imports:
  - custom_types/onudevice.yaml
description: Delete an ont
topology_template:
  node_templates:
    device#onu:
      type: tosca.nodes.ONUDevice
      properties:
        serial_number:`

/*var ontDeleteTemplate = `tosca_definitions_version: tosca_simple_yaml_1_0
imports:
  - custom_types/onudevice.yaml
description: Create a simulated OLT Device in VOLTHA
topology_template:
  node_templates:
    device#onu:
      type: something
      properties:
        serial_number:`*/

type OntDelete struct {
	ToscaDefinitionsVersion string   `yaml:"tosca_definitions_version"`
	Imports                 []string `yaml:"imports"`
	Description             string   `yaml:"description"`
	TopologyTemplate        struct {
		NodeTemplates struct {
			DeviceOnt struct {
				DeviceType string `yaml:"type"`
				Properties struct {
					SerialNumber string `yaml:"serial_number"`
				} `yaml:"properties"`
			} `yaml:"device#onu"`
		} `yaml:"node_templates"`
	} `yaml:"topology_template"`
}

func NewOntDelete(serialNumber string) OntDelete {
	o := OntDelete{}
	err := yaml.Unmarshal([]byte(ontDeleteTemplate), &o)
	if err != nil {
		fmt.Println(err)
	}
	props := &o.TopologyTemplate.NodeTemplates.DeviceOnt.Properties
	props.SerialNumber = serialNumber
	return o

}
func (ont *OntDelete) ToYaml() (string, error) {
	b, err := yaml.Marshal(ont)
	ret := string(b)
	return ret, err
}
