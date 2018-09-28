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
	"log"

	yaml "gopkg.in/yaml.v2"
)

var templateData = `
tosca_definitions_version: tosca_simple_yaml_1_0
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
           name: test
           type: test
           host: test
           port: 32
           outer_tpid: "0x8100"
           uplink: "65536"
           nas_id:
           switch_datapath_id: of:0000000000000001
           switch_port: "1"
        requirements:
          - volt_service:
              node: service#volt
              relationship: tosca.relationships.BelongsToOne
`

/*
OltProvision struct that serves as model for yaml to provsion OLT in XOS
*/

type OltProvsion struct {
	ToscaDefinitionsVersion string   `yaml:"tosca_definitions_version"`
	Imports                 []string `yaml:"imports"`
	Description             string   `yaml:"description"`
	TopologyTemplate        struct {
		NodeTemplates struct {
			ServiceVolt struct {
				Type       string `yaml:"type"`
				Properties struct {
					Name      string `yaml:"name"`
					MustExist bool   `yaml:"must-exist"`
				} `yaml:"properties"`
			} `yaml:"service#volt"`
			OltDevice struct {
				DeviceType string `yaml:"type"`
				Properties struct {
					Name             string `yaml:"name"`
					Type             string `yaml:"device_type"`
					Host             string `yaml:"host"`
					Port             int    `yaml:"port"`
					OuterTpid        string `yaml:"outer_tpid"`
					Uplink           string `yaml:"uplink"`
					NasID            string `yaml:"nas_id"`
					SwitchDataPathID string `yaml:"switch_datapath_id"`
					SwitchPort       string `yaml:"switch_port"`
				} `yaml:"properties"`
				Requirements []struct {
					VoltService struct {
						Node         string `yaml:"node"`
						Relationship string `yaml:"relationship"`
					} `yaml:"volt_service"`
				} `yaml:"requirements"`
			} `yaml:"olt_device"`
		} `yaml:"node_templates"`
	} `yaml:"topology_template"`
}

/*var templateData = `
tosca_definitions_version: tosca_simple_yaml_1_0
imports:
   - custom_types/oltdevice.yaml
   - custom_types/onudevice.yaml
   - custom_types/ponport.yaml
   - custom_types/voltservice.yaml
`

type OltProvsion struct {
	Tosca_Definitions_Version string
	Imports                   []string
}
*/

func NewOltProvision(clli string, name string, deviceType string, host string, port int) OltProvsion {
	o := OltProvsion{}
	err := yaml.Unmarshal([]byte(templateData), &o)
	if err != nil {
		log.Printf("Error un-marshalling template data %v\n", err)
	}

	props := &o.TopologyTemplate.NodeTemplates.OltDevice.Properties
	props.Name = name
	props.Type = deviceType
	props.Host = host
	props.Port = port
	props.NasID = clli
	return o
}

func (olt *OltProvsion) ToYaml() (string, error) {
	b, err := yaml.Marshal(olt)
	return string(b), err
}
