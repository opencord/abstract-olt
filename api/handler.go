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
	"fmt"
	"log"
	"net"
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
CreateChassis - allocates a new Chassis struct and stores it in chassisMap
*/
func (s *Server) CreateChassis(ctx context.Context, in *AddChassisMessage) (*AddChassisReturn, error) {
	phyChassisMap := models.GetPhyChassisMap()
	absChassisMap := models.GetAbstractChassisMap()
	clli := in.GetCLLI()
	chassis := (*phyChassisMap)[clli]
	if chassis != nil {
		return &AddChassisReturn{DeviceID: chassis.CLLI}, nil
	}
	abstractChassis := abstract.GenerateChassis(clli)
	phyChassis := &physical.Chassis{CLLI: clli, VCoreAddress: net.TCPAddr{IP: net.ParseIP(in.GetVCoreIP()), Port: int(in.GetVCorePort())}}
	if settings.GetDebug() {
		output := fmt.Sprintf("%v", abstractChassis)
		formatted := strings.Replace(output, "{", "\n{", -1)
		log.Printf("new chassis %s\n", formatted)
	}
	(*phyChassisMap)[clli] = phyChassis
	fmt.Printf("phy %v, abs %v\n", phyChassisMap, absChassisMap)
	(*absChassisMap)[clli] = abstractChassis
	return &AddChassisReturn{DeviceID: clli}, nil
}

/*
CreateOLTChassis adds an OLT chassis/line card to the Physical chassis
*/
func (s *Server) CreateOLTChassis(ctx context.Context, in *AddOLTChassisMessage) (*AddOLTChassisReturn, error) {
	fmt.Printf(" CreateOLTChassis %v \n", *in)
	phyChassisMap := models.GetPhyChassisMap()
	clli := in.GetCLLI()
	chassis := (*phyChassisMap)[clli]
	if chassis == nil {
	}
	oltType := in.GetType()
	address := net.TCPAddr{IP: net.ParseIP(in.GetSlotIP()), Port: int(in.GetSlotPort())}
	sOlt := physical.SimpleOLT{CLLI: clli, Hostname: in.GetHostname(), Address: address, Parent: chassis}

	var olt physical.OLT
	switch oltType {
	case AddOLTChassisMessage_edgecore:
		olt = physical.CreateEdgecore(&sOlt)
	case AddOLTChassisMessage_adtran:
	case AddOLTChassisMessage_tibit:
	}

	err := AddCard(chassis, olt)
	if err != nil {
		//TODO do something
	}

	return &AddOLTChassisReturn{DeviceID: in.GetHostname(), ChassisDeviceID: clli}, nil

}

/*
AddCard Adds an OLT card to an existing physical chassis, allocating ports
in the physical card to those in the abstract model
*/
func AddCard(physChassis *physical.Chassis, olt physical.OLT) error {
	physChassis.AddOLTChassis(olt)

	ports := olt.GetPorts()
	absChassis := (*models.GetAbstractChassisMap())[physChassis.CLLI]

	for i := 0; i < len(ports); i++ {
		absPort, _ := absChassis.NextPort()
		absPort.PhysPort = &ports[i]
		//AssignTraits(&ports[i], absPort)
	}
	//should probably worry about error at some point
	return nil
}

/*
ProvisionOnt provisions an ONT on a specific Chassis/LineCard/Port
*/
func (s *Server) ProvisionOnt(ctx context.Context, in *AddOntMessage) (*AddOntReturn, error) {
	absChassisMap := models.GetAbstractChassisMap()
	clli := in.GetCLLI()
	chassis := (*absChassisMap)[clli]
	err := chassis.ActivateONT(int(in.GetSlotNumber()), int(in.GetPortNumber()), int(in.GetOntNumber()), in.GetSerialNumber())

	if err != nil {
		return nil, err
	}
	return &AddOntReturn{Success: true}, nil
}

/*
DeleteOnt - deletes a previously provision ont
*/
func (s *Server) DeleteOnt(ctx context.Context, in *DeleteOntMessage) (*DeleteOntReturn, error) {
	absChassisMap := models.GetAbstractChassisMap()
	clli := in.GetCLLI()
	chassis := (*absChassisMap)[clli]
	err := chassis.DeleteONT(int(in.GetSlotNumber()), int(in.GetPortNumber()), int(in.GetOntNumber()), in.GetSerialNumber())
	if err != nil {
		return nil, err
	}
	return &DeleteOntReturn{Success: true}, nil
}
