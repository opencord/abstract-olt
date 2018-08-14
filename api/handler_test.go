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
	"golang.org/x/net/context"
)

var clli string
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
	message := &api.AddOLTChassisMessage{CLLI: clli, SlotIP: "12.2.2.0", SlotPort: 9191,
		Hostname: "SlotOne", Type: api.AddOLTChassisMessage_edgecore}
	ret, err := server.CreateOLTChassis(ctx, message)
	if err != nil {
		t.Fatalf("CreateOLTChassis failed %v\n", err)
	}
	fmt.Printf("CreateOLTChassis success %v\n", ret)

}
