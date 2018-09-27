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

package physical_test

import (
	"net"
	"testing"

	"gerrit.opencord.org/abstract-olt/models/physical"
)

func TestPhysical_Edgecore(t *testing.T) {
	clli := "test_clli"
	chassis := physical.Chassis{CLLI: clli}
	hostname := "my_name"
	ip := "192.168.0.1"
	port := 33
	address := net.TCPAddr{IP: net.ParseIP(ip), Port: port}
	parent := &chassis
	switchPort := 3

	olt := &physical.SimpleOLT{CLLI: clli, Hostname: hostname, Address: address, Parent: parent, DataSwitchPort: switchPort}
	olt.CreateEdgecore()
	if olt.GetCLLI() != clli {
		t.Fatal("Failed to assign CLLI in CreateEdgecore")
	}
}
