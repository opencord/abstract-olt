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

package api

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"gerrit.opencord.org/abstract-olt/internal/pkg/impl"
	"gerrit.opencord.org/abstract-olt/models/inventory"
	context "golang.org/x/net/context"
)

/*
Server instance of the grpc server
*/
type Server struct {
}

var syncChan chan bool
var isDirty bool

var runOnce sync.Once

func getSyncChannel() chan bool {
	runOnce.Do(func() {
		syncChan = make(chan bool, 1)
		syncChan <- true
	})

	return syncChan
}
func done(myChan chan bool, done bool) {
	myChan <- done
}

/*
Echo - Tester function which just returns same string sent to it
*/
func (s *Server) Echo(ctx context.Context, in *EchoMessage) (*EchoReplyMessage, error) {
	myChan := getSyncChannel()
	<-myChan
	defer done(myChan, true)
	ping := in.GetPing()
	pong := EchoReplyMessage{Pong: ping}
	return &pong, nil
}

/*
CreateChassis - allocates a new Chassis struct and stores it in chassisMap
*/
func (s *Server) CreateChassis(ctx context.Context, in *AddChassisMessage) (*AddChassisReturn, error) {
	clli := in.GetCLLI()

	xosIP := net.ParseIP(in.GetXOSIP())
	xosPort := int(in.GetXOSPort())
	if xosIP == nil {
		errStr := fmt.Sprintf("Invalid IP %s supplied for XOSIP", in.GetXOSIP())
		return nil, errors.New(errStr)
	}
	xosAddress := net.TCPAddr{IP: xosIP, Port: xosPort}
	xosUser := in.GetXOSUser()
	xosPassword := in.GetXOSPassword()
	if xosUser == "" || xosPassword == "" {
		return nil, errors.New("Either XOSUser or XOSPassword supplied were empty")
	}
	shelf := int(in.GetShelf())
	rack := int(in.GetRack())
	deviceID, err := impl.CreateChassis(clli, xosAddress, xosUser, xosPassword, shelf, rack)
	if err != nil {
		return nil, err
	}
	return &AddChassisReturn{DeviceID: deviceID}, nil
}

/*
ChangeXOSUserPassword - allows update of xos credentials
*/
func (s *Server) ChangeXOSUserPassword(ctx context.Context, in *ChangeXOSUserPasswordMessage) (*ChangeXOSUserPasswordReturn, error) {
	clli := in.GetCLLI()
	xosUser := in.GetXOSUser()
	xosPassword := in.GetXOSPassword()
	if xosUser == "" || xosPassword == "" {
		return nil, errors.New("Either XOSUser or XOSPassword supplied were empty")
	}
	success, err := impl.ChangeXOSUserPassword(clli, xosUser, xosPassword)
	return &ChangeXOSUserPasswordReturn{Success: success}, err

}

/*
CreateOLTChassis adds an OLT chassis/line card to the Physical chassis
*/
func (s *Server) CreateOLTChassis(ctx context.Context, in *AddOLTChassisMessage) (*AddOLTChassisReturn, error) {
	clli := in.GetCLLI()
	oltType := in.GetType().String()
	address := net.TCPAddr{IP: net.ParseIP(in.GetSlotIP()), Port: int(in.GetSlotPort())}
	hostname := in.GetHostname()
	clli, err := impl.CreateOLTChassis(clli, oltType, address, hostname)
	return &AddOLTChassisReturn{DeviceID: hostname, ChassisDeviceID: clli}, err
}

/*
ProvisionOnt provisions an ONT on a specific Chassis/LineCard/Port
*/
func (s *Server) ProvisionOnt(ctx context.Context, in *AddOntMessage) (*AddOntReturn, error) {
	clli := in.GetCLLI()
	slotNumber := int(in.GetSlotNumber())
	portNumber := int(in.GetPortNumber())
	ontNumber := int(in.GetOntNumber())
	serialNumber := in.GetSerialNumber()
	success, err := impl.ProvisionOnt(clli, slotNumber, portNumber, ontNumber, serialNumber)
	return &AddOntReturn{Success: success}, err
}

/*
ProvisionOntFull - provisions ont using sTag,cTag,NasPortID, and CircuitID passed in
*/
func (s *Server) ProvisionOntFull(ctx context.Context, in *AddOntFullMessage) (*AddOntReturn, error) {
	clli := in.GetCLLI()
	slotNumber := int(in.GetSlotNumber())
	portNumber := int(in.GetPortNumber())
	ontNumber := int(in.GetOntNumber())
	serialNumber := in.GetSerialNumber()
	cTag := in.GetCTag()
	sTag := in.GetSTag()
	nasPortID := in.GetNasPortID()
	circuitID := in.GetCircuitID()
	success, err := impl.ProvisionOntFull(clli, slotNumber, portNumber, ontNumber, serialNumber, cTag, sTag, nasPortID, circuitID)
	return &AddOntReturn{Success: success}, err
}

/*
DeleteOnt - deletes a previously provision ont
*/
func (s *Server) DeleteOnt(ctx context.Context, in *DeleteOntMessage) (*DeleteOntReturn, error) {
	clli := in.GetCLLI()
	slotNumber := int(in.GetSlotNumber())
	portNumber := int(in.GetPortNumber())
	ontNumber := int(in.GetOntNumber())
	serialNumber := in.GetSerialNumber()
	success, err := impl.DeleteOnt(clli, slotNumber, portNumber, ontNumber, serialNumber)
	return &DeleteOntReturn{Success: success}, err
}

/*
Reflow - iterates through provisioning to rebuild Seba-Pod
*/
func (s *Server) Reflow(ctx context.Context, in *ReflowMessage) (*ReflowReturn, error) {
	success, err := impl.Reflow()
	return &ReflowReturn{Success: success}, err

}

/*
Output - causes an immediate backup to be created
*/
func (s *Server) Output(ctx context.Context, in *OutputMessage) (*OutputReturn, error) {
	success, err := impl.DoOutput()
	return &OutputReturn{Success: success}, err

}

/*
GetFullInventory - gets a full json dump of the currently provisioned equipment
*/
func (s *Server) GetFullInventory(ctx context.Context, in *FullInventoryMessage) (*InventoryReturn, error) {
	json := inventory.GatherAllInventory()
	return &InventoryReturn{JsonDump: json}, nil
}

/*
GetInventory - returns a json dump of a particular seba-pod
*/
func (s *Server) GetInventory(ctx context.Context, in *InventoryMessage) (*InventoryReturn, error) {
	json, err := inventory.GatherInventory(in.GetClli())
	return &InventoryReturn{JsonDump: json}, err
}
