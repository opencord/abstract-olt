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

package models_test

import (
	"net"
	"testing"

	"gerrit.opencord.org/abstract-olt/internal/pkg/settings"
	"gerrit.opencord.org/abstract-olt/models"
	"gerrit.opencord.org/abstract-olt/models/abstract"
	"gerrit.opencord.org/abstract-olt/models/physical"
)

var chassisMap *map[string]*models.ChassisHolder
var clli string
var json []byte

func TestChassisSerialize_Serialize(t *testing.T) {
	//func (chassisHolder ChassisHolder) Serialize() ([]byte, error) {
	settings.SetDummy(true)
	clli = "TEST_CLLI"
	chassisMap = models.GetChassisMap()
	abstractChassis := abstract.GenerateChassis(clli, 1, 1)
	phyChassis := physical.Chassis{CLLI: clli, XOSAddress: net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}, Rack: 1, Shelf: 1}
	chassisHolder := &models.ChassisHolder{AbstractChassis: abstractChassis, PhysicalChassis: phyChassis}
	(*chassisMap)[clli] = chassisHolder
	sOlt := physical.SimpleOLT{CLLI: clli, Hostname: "slot1", Address: net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}, Parent: &phyChassis}
	phyChassis.AddOLTChassis(sOlt)
	ports := sOlt.GetPorts()
	for i := 0; i < len(ports); i++ {
		absPort, _ := chassisHolder.AbstractChassis.NextPort()

		absPort.PhysPort = &ports[i]
		//AssignTraits(&ports[i], absPort)
	}
	var err error
	json, err = chassisHolder.Serialize()
	if err != nil {
		t.Fatalf("TestChassisSerialize_Serialize failed with %v\n", err)
	}
}

func TestChassisSerialize_Deserialize(t *testing.T) {
	//func (chassisHolder *ChassisHolder) Deserialize(jsonData []byte) error {
	chassisHolder := models.ChassisHolder{}
	err := chassisHolder.Deserialize(json)
	if err != nil {
		t.Fatalf("Deserialize threw an error %v\n", err)
	}
	newJSON, _ := chassisHolder.Serialize()
	if string(json) != string(newJSON) {
		t.Fatalf("Failed to de-serialize and serialize accurately")
	}
}
