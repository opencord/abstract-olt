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
	"strings"

	yaml "gopkg.in/yaml.v2"
)

var subSubscriberTemplate = `tosca_definitions_version: tosca_simple_yaml_1_0
imports:
  - custom_types/rcordsubscriber.yaml

description: Pre-provsion a subscriber
topology_template:
  node_templates:
    RG_NAME:
      type: tosca.nodes.RCORDSubscriber
      properties:
        name:
        status: pre-provisioned
        c_tag:
        s_tag:
        onu_device:
        nas_port_id:
        circuit_id:
        remote_id:`

type SubscriberProvision struct {
	ToscaDefinitionsVersion string   `yaml:"tosca_definitions_version"`
	Imports                 []string `yaml:"imports"`
	Description             string   `yaml:"description"`
	TopologyTemplate        struct {
		NodeTemplates struct {
			RgName struct {
				Type       string `yaml:"type`
				Properties struct {
					Name      string `yaml:"name"`
					Status    string `yaml:"status"`
					CTag      uint32 `yaml:"c_tag"`
					STag      uint32 `yaml:"s_tag"`
					OnuDevice string `yaml:"onu_device"`
					NasPortID string `yaml:"nas_port_id"`
					CircuitID string `yaml:"circuit_id"`
					RemoteID  string `yaml:"remote_id"`
				} `yaml:"properties"`
			} `yaml:"RG_NAME"`
		} `yaml:"node_templates"`
	} `yaml:"topology_template"`
}

func NewSubscriberProvision(name string, cTag uint32, sTag uint32, onuDevice string, nasPortID string, circuitID string, remoteID string) SubscriberProvision {
	s := SubscriberProvision{}
	err := yaml.Unmarshal([]byte(subSubscriberTemplate), &s)
	if err != nil {
		log.Printf("Error un-marshalling template data %v\n", err)
	}
	props := &s.TopologyTemplate.NodeTemplates.RgName.Properties
	props.Name = name
	props.CTag = cTag
	props.STag = sTag
	props.OnuDevice = onuDevice
	props.NasPortID = nasPortID
	props.CircuitID = circuitID
	props.RemoteID = remoteID
	return s
}
func (sub *SubscriberProvision) ToYaml() (string, error) {
	b, err := yaml.Marshal(sub)
	ret := string(b)
	name := sub.TopologyTemplate.NodeTemplates.RgName.Properties.Name
	ret = strings.Replace(ret, "RG_NAME", name, -1)
	return ret, err
}
