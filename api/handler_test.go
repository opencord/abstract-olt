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

package api_test

import (
	"fmt"
	"testing"

	"gerrit.opencord.org/abstract-olt/api"
	"gerrit.opencord.org/abstract-olt/models/abstract"
	"gerrit.opencord.org/abstract-olt/models/physical"
	"golang.org/x/net/context"
)

var clli string
var slotHostname = "SlotOne"
var ctx context.Context
var server api.Server

func TestHandler_CreateChassis(t *testing.T) {
	fmt.Println("in handlerTest_CreateChassis")
	ctx = context.Background()
	server = api.Server{}
	message := &api.AddChassisMessage{CLLI: "my cilli", VCoreIP: "192.168.0.1", VCorePort: 9191}
	ret, err := server.CreateChassis(ctx, message)
	if err != nil {
		t.Fatalf("CreateChassis failed %v\n", err)
	}
	clli = ret.DeviceID
}
func TestHandler_CreateOLTChassis(t *testing.T) {
	fmt.Println("in handlerTest_CreateChassis")
	message := &api.AddOLTChassisMessage{CLLI: clli, SlotIP: "12.2.2.0", SlotPort: 9191,
		Hostname: slotHostname, Type: api.AddOLTChassisMessage_edgecore}
	ret, err := server.CreateOLTChassis(ctx, message)
	if err != nil {
		t.Fatalf("CreateOLTChassis failed %v\n", err)
	}
	fmt.Printf("CreateOLTChassis success %v\n", ret)
}
func TestHandler_EnableSlot(t *testing.T) {
	ctx = context.Background()
	server = api.Server{}
	fmt.Println("in handler_test_EnableSlot")
	// slot number 1 should be provisioned above
	message := &api.ActivateSlotMessage{CLLI: clli, SlotNumber: 1}
	ret, err := server.EnableSlot(ctx, message)
	if err != nil {
		t.Fatalf("EnableSlot failed with %v\n", err)
	}
	fmt.Printf("EnableSlot succeeded with %v\n", ret)
	// Slot 2 isn't provisioned and should fail
	message = &api.ActivateSlotMessage{CLLI: clli, SlotNumber: 2}
	ret, err = server.EnableSlot(ctx, message)
	if err != nil {
		switch err.(type) {
		case *physical.UnprovisionedSlotError:
			fmt.Printf("EnableSlot failed as it should with %v\n", err)
		default:
			t.Fatalf("EnableSlot failed with %v\n", err)
		}
		t.Fatalf("EnableSlot should have failed but didn't")
	}

}
func TestHandler_ProvisionOnt(t *testing.T) {
	ctx = context.Background()
	server = api.Server{}
	fmt.Println("in handlerTest_CreateChassis")
	// this one should succeed
	message := &api.AddOntMessage{CLLI: clli, SlotNumber: 1, PortNumber: 3, OntNumber: 2, SerialNumber: "2033029402"}
	ret, err := server.ProvisionOnt(ctx, message)
	if err != nil {
		t.Fatalf("ProvisionOnt failed %v\n", err)
	}
	// this one should fail
	fmt.Println("here")
	//SlotNumber 1 hasn't been provisioned
	message = &api.AddOntMessage{CLLI: clli, SlotNumber: 2, PortNumber: 3, OntNumber: 2, SerialNumber: "2033029402"}
	ret, err = server.ProvisionOnt(ctx, message)
	if err != nil {
		switch err.(type) {
		case *abstract.UnprovisonedPortError:
			fmt.Printf("ProvisionOnt failed as it should with %v\n", err)
		default:
			t.Fatalf("ProvsionOnt test failed with %v\n", err)
		}
	} else {
		t.Fatalf("ProvsionOnt should have failed but didn't")
	}
	fmt.Printf("ProvisionOnt success %v\n", ret)
}
