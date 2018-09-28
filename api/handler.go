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
	"log"
	"net"
	"os"
	"strings"

	"gerrit.opencord.org/abstract-olt/internal/pkg/settings"
	"gerrit.opencord.org/abstract-olt/models"
	"gerrit.opencord.org/abstract-olt/models/abstract"
	"gerrit.opencord.org/abstract-olt/models/physical"
	context "golang.org/x/net/context"
)

/*
Server instance of the grpc server
*/
type Server struct {
}

/*
Echo - Tester function which just returns same string sent to it
*/
func (s *Server) Echo(ctx context.Context, in *EchoMessage) (*EchoReplyMessage, error) {
	ping := in.GetPing()
	pong := EchoReplyMessage{Pong: ping}
	return &pong, nil
}

/*
CreateChassis - allocates a new Chassis struct and stores it in chassisMap
*/
func (s *Server) CreateChassis(ctx context.Context, in *AddChassisMessage) (*AddChassisReturn, error) {
	chassisMap := models.GetChassisMap()
	clli := in.GetCLLI()
	chassisHolder := (*chassisMap)[clli]
	if chassisHolder != nil {
		return &AddChassisReturn{DeviceID: chassisHolder.AbstractChassis.CLLI}, nil
	}

	abstractChassis := abstract.GenerateChassis(clli, int(in.GetRack()), int(in.GetShelf()))
	phyChassis := physical.Chassis{CLLI: clli, VCoreAddress: net.TCPAddr{IP: net.ParseIP(in.GetVCoreIP()),
		Port: int(in.GetVCorePort())}, Rack: int(in.GetRack()), Shelf: int(in.GetShelf())}

	chassisHolder = &models.ChassisHolder{AbstractChassis: abstractChassis, PhysicalChassis: phyChassis}
	if settings.GetDebug() {
		output := fmt.Sprintf("%v", abstractChassis)
		formatted := strings.Replace(output, "{", "\n{", -1)
		log.Printf("new chassis %s\n", formatted)
	}
	(*chassisMap)[clli] = chassisHolder
	return &AddChassisReturn{DeviceID: clli}, nil
}

/*
CreateOLTChassis adds an OLT chassis/line card to the Physical chassis
*/
func (s *Server) CreateOLTChassis(ctx context.Context, in *AddOLTChassisMessage) (*AddOLTChassisReturn, error) {
	fmt.Printf(" CreateOLTChassis %v \n", *in)
	chassisMap := models.GetChassisMap()
	clli := in.GetCLLI()
	chassisHolder := (*chassisMap)[clli]
	if chassisHolder == nil {
		errString := fmt.Sprintf("There is no chassis with CLLI of %s", clli)
		return &AddOLTChassisReturn{DeviceID: "", ChassisDeviceID: ""}, errors.New(errString)
	}
	oltType := in.GetType()
	address := net.TCPAddr{IP: net.ParseIP(in.GetSlotIP()), Port: int(in.GetSlotPort())}
	physicalChassis := &chassisHolder.PhysicalChassis
	sOlt := physical.SimpleOLT{CLLI: clli, Hostname: in.GetHostname(), Address: address, Parent: physicalChassis}
	switch oltType {
	case AddOLTChassisMessage_edgecore:
		sOlt.CreateEdgecore()
	case AddOLTChassisMessage_adtran:
	case AddOLTChassisMessage_tibit:
	}
	physicalChassis.AddOLTChassis(sOlt)
	ports := sOlt.GetPorts()
	for i := 0; i < len(ports); i++ {
		absPort, _ := chassisHolder.AbstractChassis.NextPort()

		absPort.PhysPort = &ports[i]
		//AssignTraits(&ports[i], absPort)
	}
	return &AddOLTChassisReturn{DeviceID: in.GetHostname(), ChassisDeviceID: clli}, nil

}

/*
ProvisionOnt provisions an ONT on a specific Chassis/LineCard/Port
*/
func (s *Server) ProvisionOnt(ctx context.Context, in *AddOntMessage) (*AddOntReturn, error) {
	chassisMap := models.GetChassisMap()
	clli := in.GetCLLI()
	chassisHolder := (*chassisMap)[clli]
	if chassisHolder == nil {
		errString := fmt.Sprintf("There is no chassis with CLLI of %s", clli)
		return &AddOntReturn{Success: false}, errors.New(errString)
	}
	err := chassisHolder.AbstractChassis.ActivateONT(int(in.GetSlotNumber()), int(in.GetPortNumber()), int(in.GetOntNumber()), in.GetSerialNumber())

	if err != nil {
		return nil, err
	}
	return &AddOntReturn{Success: true}, nil
}

/*
DeleteOnt - deletes a previously provision ont
*/
func (s *Server) DeleteOnt(ctx context.Context, in *DeleteOntMessage) (*DeleteOntReturn, error) {
	chassisMap := models.GetChassisMap()
	clli := in.GetCLLI()
	chassisHolder := (*chassisMap)[clli]
	err := chassisHolder.AbstractChassis.DeleteONT(int(in.GetSlotNumber()), int(in.GetPortNumber()), int(in.GetOntNumber()), in.GetSerialNumber())
	if err != nil {
		return nil, err
	}
	return &DeleteOntReturn{Success: true}, nil
}

func (s *Server) Output(ctx context.Context, in *OutputMessage) (*OutputReturn, error) {
	chassisMap := models.GetChassisMap()
	for clli, chassisHolder := range *chassisMap {

		json, _ := (chassisHolder).Serialize()
		backupFile := fmt.Sprintf("backup/%s", clli)
		f, _ := os.Create(backupFile)
		defer f.Close()
		_, _ = f.WriteString(string(json))
		f.Sync()
	}
	return &OutputReturn{Success: true}, nil

}
