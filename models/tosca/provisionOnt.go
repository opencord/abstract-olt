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
	"net"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

var ontTemplate = `tosca_definitions_version: tosca_simple_yaml_1_0
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
    ont:
      type: tosca.nodes.AttWorkflowDriverWhiteListEntry
      properties:
        serial_number: ALPHe3d1cf57
        pon_port_id: 536870912
        device_id: of:000000000a4001ce
      requirements:
        - owner:
            node: service#att
            relationship: tosca.relationships.BelongsToOne`

type OntProvision struct {
	ToscaDefinitionsVersion string   `yaml:"tosca_definitions_version"`
	Imports                 []string `yaml:"imports"`
	Description             string   `yaml:"description"`
	TopologyTemplate        struct {
		NodeTemplates struct {
			ServiceATT struct {
				Type       string `yaml:"type"`
				Properties struct {
					Name      string `yaml:"name"`
					MustExist bool   `yaml:"must-exist"`
				} `yaml:"properties"`
			} `yaml:"service#att"`
			Ont struct {
				DeviceType string `yaml:"type"`
				Properties struct {
					SerialNumber string `yaml:"serial_number"`
					PonPortID    int    `yaml:"pon_port_id"`
					DeviceID     string `yaml:"device_id"`
				} `yaml:"properties"`
				Requirements []struct {
					Owner struct {
						Node         string `yaml:"node"`
						Relationship string `yaml:"relationship"`
					} `yaml:"owner"`
				} `yaml:"requirements"`
			} `yaml:"ont"`
		} `yaml:"node_templates"`
	} `yaml:"topology_template"`
}

func NewOntProvision(serialNumber string, oltIP net.IP, ponPortNumber int) OntProvision {
	offset := 1 << 29
	o := OntProvision{}
	err := yaml.Unmarshal([]byte(ontTemplate), &o)
	if err != nil {
	}
	props := &o.TopologyTemplate.NodeTemplates.Ont.Properties
	props.PonPortID = offset + (ponPortNumber - 1)
	props.SerialNumber = serialNumber
	ipNum := []byte(oltIP[12:16]) //only handling ipv4
	ofID := fmt.Sprintf("of:00000000%0x", ipNum)
	props.DeviceID = ofID
	return o

}
func (ont *OntProvision) ToYaml() (string, error) {
	b, err := yaml.Marshal(ont)
	ret := string(b)
	// Damn dirty hack but what are you going to do????
	serialNumber := ont.TopologyTemplate.NodeTemplates.Ont.Properties.SerialNumber
	ret = strings.Replace(ret, "ont", serialNumber, -1)
	return ret, err
}
