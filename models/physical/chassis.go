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

package physical

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"gerrit.opencord.org/abstract-olt/contrib/xos"
	"gerrit.opencord.org/abstract-olt/internal/pkg/settings"
	"gerrit.opencord.org/abstract-olt/models/tosca"
	"google.golang.org/grpc"
)

type basicAuth struct {
	username string
	password string
}

func (b basicAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	auth := b.username + ":" + b.password
	enc := base64.StdEncoding.EncodeToString([]byte(auth))
	return map[string]string{
		"authorization": "Basic " + enc,
	}, nil
}

func (basicAuth) RequireTransportSecurity() bool {
	return false
}

/*
Chassis is a model that takes up to 16 discreet OLT chassis as if it is a 16 slot OLT chassis
*/
type Chassis struct {
	CLLI        string
	XOSAddress  net.TCPAddr
	XOSUser     string
	XOSPassword string
	Linecards   []SimpleOLT
	Rack        int
	Shelf       int
}

/*
UnprovisionedSlotError - Error thrown when attempting to provision to a line card that hasn't been activated
*/
type UnprovisionedSlotError struct {
	CLLI       string
	SlotNumber int
}

func (e *UnprovisionedSlotError) Error() string {
	return fmt.Sprintf("SlotNumber %d in Chassis %s is currently unprovsioned", e.SlotNumber, e.CLLI)
}

/*
AddOLTChassis - adds a reference to a new olt chassis
*/
func (chassis *Chassis) AddOLTChassis(olt SimpleOLT) {
	olt.SetNumber((len(chassis.Linecards) + 1))
	chassis.Linecards = append(chassis.Linecards, olt)
	if settings.GetGrpc() {
		chassis.SendOltGRPC(olt)
	} else {
		chassis.SendOltTosca(olt)
	}
}

/*
SendOltGRPC - provisions olt using grpc interface
*/
func (chassis *Chassis) SendOltGRPC(olt SimpleOLT) error {
	if settings.GetDummy() {
		log.Println("Running in Dummy mode with GRPC in SendOltGRPC")
		return nil
	}
	conn, err := grpc.Dial(chassis.XOSAddress.String(), grpc.WithInsecure(), grpc.WithPerRPCCredentials(basicAuth{
		username: chassis.XOSUser,
		password: chassis.XOSPassword,
	}))
	defer conn.Close()
	if err != nil {
		log.Println(err)
		return err
	}

	xosClient := xos.NewXosClient(conn)
	//queryElement := &xos.QueryElement{Operator: xos.QueryElement_EQUAL, Name: "volt_service_instances", Value: &xos.QueryElement_SValue{"volt"}}
	queryElement := &xos.QueryElement{Operator: xos.QueryElement_EQUAL, Name: "name", Value: &xos.QueryElement_SValue{"volt"}}
	queryElements := []*xos.QueryElement{queryElement}
	query := &xos.Query{Kind: xos.Query_DEFAULT, Elements: queryElements}

	voltResponse, err := xosClient.FilterVOLTService(context.Background(), query)
	if err != nil {
		log.Println(err)
		return err
	}
	voltServices := voltResponse.GetItems()
	if len(voltServices) == 0 {
		return errors.New("xosClient.FilterVOLTService returned 0 entries with name \"volt\"")
	}
	voltService := voltServices[0]

	response, err := xosClient.CreateOLTDevice(context.Background(), &xos.OLTDevice{
		NamePresent:             &xos.OLTDevice_Name{olt.Hostname},
		DeviceTypePresent:       &xos.OLTDevice_DeviceType{olt.Driver},
		HostPresent:             &xos.OLTDevice_Host{olt.GetAddress().IP.String()},
		PortPresent:             &xos.OLTDevice_Port{int32(olt.GetAddress().Port)},
		OuterTpidPresent:        &xos.OLTDevice_OuterTpid{"0x8100"},
		UplinkPresent:           &xos.OLTDevice_Uplink{"65536"},
		NasIdPresent:            &xos.OLTDevice_NasId{olt.CLLI},
		SwitchDatapathIdPresent: &xos.OLTDevice_SwitchDatapathId{"of:0000000000000001"},
		SwitchPortPresent:       &xos.OLTDevice_SwitchPort{"1"},
		VoltServicePresent:      &xos.OLTDevice_VoltServiceId{voltService.GetId()},
	})
	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		return err
	}
	log.Printf("Response is %v\n", response)
	return nil
}

/*
SendOltTosca - Provision OLT using TOSCA Interface
*/
func (chassis *Chassis) SendOltTosca(olt SimpleOLT) error {
	ipString := olt.GetAddress().IP.String()
	webServerPort := olt.GetAddress().Port
	oltStruct := tosca.NewOltProvision(chassis.CLLI, olt.GetHostname(), olt.Driver, ipString, webServerPort)
	yaml, _ := oltStruct.ToYaml()
	if settings.GetDummy() {
		log.Printf("yaml:%s\n", yaml)
		log.Println("YAML IS NOT BEING SET TO XOS or DEBUG is Set")
		return nil
	}

	if settings.GetDebug() {
		log.Printf("yaml:%s\n", yaml)
	}
	client := &http.Client{}
	requestList := fmt.Sprintf("http://%s:%d/run", chassis.XOSAddress.IP.String(), chassis.XOSAddress.Port)
	req, err := http.NewRequest("POST", requestList, strings.NewReader(yaml))
	req.Header.Add("xos-username", chassis.XOSUser)
	req.Header.Add("xos-password", chassis.XOSPassword)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		return err
	}
	log.Printf("Server response was %v\n", resp.Body)
	return nil
}
func (chassis *Chassis) provisionONT(ont Ont) {
	//TODO - api call to provison s/c vlans and ont serial number etc
	log.Printf("chassis.provisionONT(%s,SVlan:%d,CVlan:%d)\n", ont.SerialNumber, ont.Svlan, ont.Cvlan)
	if settings.GetGrpc() {
		chassis.SendOntGRPC(ont)
		chassis.SendSubscriberGRPC(ont)
	} else {
		chassis.SendOntTosca(ont)
		chassis.SendSubscriberTosca(ont)
	}
}

/*
SendOntGRPC - Provision ONT on XOS using GRPC interface
*/
func (chassis *Chassis) SendOntGRPC(ont Ont) error {
	if settings.GetDummy() {
		log.Println("Running in Dummy mode with GRPC in SendOntGRPC")
		return nil
	}
	ponPort := ont.Parent
	slot := ponPort.Parent
	ip := slot.Address.IP
	ipNum := []byte(ip[12:16]) //only handling ipv4
	ofID := fmt.Sprintf("of:00000000%0x", ipNum)

	conn, err := grpc.Dial(chassis.XOSAddress.String(), grpc.WithInsecure(), grpc.WithPerRPCCredentials(basicAuth{
		username: chassis.XOSUser,
		password: chassis.XOSPassword,
	},
	))
	defer conn.Close()
	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		return err
	}
	xosClient := xos.NewXosClient(conn)
	queryElement := &xos.QueryElement{Operator: xos.QueryElement_EQUAL, Name: "name", Value: &xos.QueryElement_SValue{"att-workflow-driver"}}
	queryElements := []*xos.QueryElement{queryElement}
	query := &xos.Query{Kind: xos.Query_DEFAULT, Elements: queryElements}

	attWorkFlowResponse, err := xosClient.FilterAttWorkflowDriverService(context.Background(), query)
	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		return err
	}

	attWorkFlowServices := attWorkFlowResponse.GetItems()
	if len(attWorkFlowServices) == 0 {
		err := "xosClient.FilterAttWorkflowDriverService return zero attWorkFlowServices with name att-workflow-driver"
		log.Print(err)
		return errors.New(err)
	}
	log.Printf("FilterAttWorkflowDriver response is %s", attWorkFlowResponse)
	attWorkFlowService := attWorkFlowServices[0]

	newQueryElement := &xos.QueryElement{Operator: xos.QueryElement_EQUAL, Name: "serial_number", Value: &xos.QueryElement_SValue{ont.SerialNumber}}
	newQueryElements := []*xos.QueryElement{newQueryElement}
	query = &xos.Query{Kind: xos.Query_DEFAULT, Elements: newQueryElements}

	onuResponse, err := xosClient.FilterONUDevice(context.Background(), query)
	log.Printf("FilterONU response is %s", onuResponse)
	onus := onuResponse.GetItems()
	if len(onus) == 0 {
		err := fmt.Sprintf("xosClient.FilterONUDevices return zero onus with serial number %s", ont.SerialNumber)
		log.Print(err)
		return errors.New(err)
	}
	deviceID := onus[0].GetDeviceId()

	offset := 1 << 29
	ponPortNumber := offset + (ponPort.Number - 1)
	log.Printf("Calling xosClient.CreateAttWorkflowDriverWhiteListEntry with SerialNumberPresent: %s DeviceIdPresent: %s PonPortIdPresent: %d OwnerPresent: %d", ont.SerialNumber, deviceID, ponPortNumber, attWorkFlowService.GetId())
	response, err := xosClient.CreateAttWorkflowDriverWhiteListEntry(context.Background(), &xos.AttWorkflowDriverWhiteListEntry{
		SerialNumberPresent: &xos.AttWorkflowDriverWhiteListEntry_SerialNumber{ont.SerialNumber},
		//DeviceIdPresent:     &xos.AttWorkflowDriverWhiteListEntry_DeviceId{deviceID},
		DeviceIdPresent:  &xos.AttWorkflowDriverWhiteListEntry_DeviceId{ofID},
		PonPortIdPresent: &xos.AttWorkflowDriverWhiteListEntry_PonPortId{int32(ponPortNumber)},
		OwnerPresent:     &xos.AttWorkflowDriverWhiteListEntry_OwnerId{attWorkFlowService.GetId()},
	})

	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		return err
	}
	log.Printf("Response is %v\n", response)
	return nil
}

/*
SendOntTosca - Provision ONT on XOS using Tosca interface
*/
func (chassis *Chassis) SendOntTosca(ont Ont) error {
	ponPort := ont.Parent
	slot := ponPort.Parent
	ontStruct := tosca.NewOntProvision(ont.SerialNumber, slot.Address.IP, ponPort.Number)
	yaml, _ := ontStruct.ToYaml()

	if settings.GetDummy() {
		log.Printf("yaml:%s\n", yaml)
		log.Println("YAML IS NOT BEING SET TO XOS")
		return nil
	}
	client := &http.Client{}
	requestList := fmt.Sprintf("http://%s:%d/run", chassis.XOSAddress.IP.String(), chassis.XOSAddress.Port)
	req, err := http.NewRequest("POST", requestList, strings.NewReader(yaml))
	req.Header.Add("xos-username", chassis.XOSUser)
	req.Header.Add("xos-password", chassis.XOSPassword)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		return err
	}
	log.Printf("Response is %v\n", resp)
	return nil
}

/*
SendSubscriberGRPC - Provisons a subscriber using the GRPC Interface
*/
func (chassis *Chassis) SendSubscriberGRPC(ont Ont) error {
	if settings.GetDummy() {
		log.Println("Running in Dummy mode with GRPC in SendSubscriberGRPC")
		return nil
	}
	ponPort := ont.Parent
	slot := ponPort.Parent
	conn, err := grpc.Dial(chassis.XOSAddress.String(), grpc.WithInsecure(), grpc.WithPerRPCCredentials(basicAuth{
		username: chassis.XOSUser,
		password: chassis.XOSPassword,
	}))
	defer conn.Close()
	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		return err
	}

	xosClient := xos.NewXosClient(conn)
	rgName := fmt.Sprintf("%s_%d_%d_%d_RG", chassis.CLLI, slot.Number, ponPort.Number, ont.Number)
	response, err := xosClient.CreateRCORDSubscriber(context.Background(), &xos.RCORDSubscriber{
		NamePresent:      &xos.RCORDSubscriber_Name{rgName},
		CTagPresent:      &xos.RCORDSubscriber_CTag{int32(ont.Cvlan)},
		STagPresent:      &xos.RCORDSubscriber_STag{int32(ont.Svlan)},
		OnuDevicePresent: &xos.RCORDSubscriber_OnuDevice{ont.SerialNumber},
		NasPortIdPresent: &xos.RCORDSubscriber_NasPortId{ont.NasPortID},
		CircuitIdPresent: &xos.RCORDSubscriber_CircuitId{ont.CircuitID},
		RemoteIdPresent:  &xos.RCORDSubscriber_RemoteId{chassis.CLLI}})
	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		return err
	}
	log.Println(response)
	return nil

}

/*
SendSubscriberTosca - Provisons a subscriber using the Tosca Interface
*/
func (chassis *Chassis) SendSubscriberTosca(ont Ont) error {
	ponPort := ont.Parent
	slot := ponPort.Parent
	requestList := fmt.Sprintf("http://%s:%d/run", chassis.XOSAddress.IP.String(), chassis.XOSAddress.Port)
	rgName := fmt.Sprintf("%s_%d_%d_%d_RG", chassis.CLLI, slot.Number, ponPort.Number, ont.Number)
	subStruct := tosca.NewSubscriberProvision(rgName, ont.Cvlan, ont.Svlan, ont.SerialNumber, ont.NasPortID, ont.CircuitID, chassis.CLLI)
	yaml, _ := subStruct.ToYaml()
	if settings.GetDummy() {
		log.Printf("yaml:%s\n", yaml)
		log.Println("YAML IS NOT BEING SET TO XOS")
		return nil
	}
	req, err := http.NewRequest("POST", requestList, strings.NewReader(yaml))
	req.Header.Add("xos-username", chassis.XOSUser)
	req.Header.Add("xos-password", chassis.XOSPassword)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		return err
	}
	log.Printf("Response is %v\n", resp)

	return nil
}

func (chassis *Chassis) deleteONT(ont Ont) {
	log.Printf("chassis.deleteONT(%s,SVlan:%d,CVlan:%d)\n", ont.SerialNumber, ont.Svlan, ont.Cvlan)
	if settings.GetGrpc() {
		chassis.deleteOntWhitelistGRPC(ont)
	} else {
		chassis.deleteOntTosca(ont)
	}
}

/*
deleteOntGRPC - deletes ONT using XOS GRPC Interface
*/
func (chassis *Chassis) deleteOntWhitelistGRPC(ont Ont) error {
	if settings.GetDummy() {
		log.Println("Running in Dummy mode with GRPC in SendSubscriberGRPC")
		return nil
	}
	conn, err := grpc.Dial(chassis.XOSAddress.String(), grpc.WithInsecure(), grpc.WithPerRPCCredentials(basicAuth{
		username: chassis.XOSUser,
		password: chassis.XOSPassword,
	}))
	defer conn.Close()
	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		return err
	}
	xosClient := xos.NewXosClient(conn)
	queryElement := &xos.QueryElement{Operator: xos.QueryElement_EQUAL, Name: "serial_number", Value: &xos.QueryElement_SValue{ont.SerialNumber}}
	queryElements := []*xos.QueryElement{queryElement}
	query := &xos.Query{Kind: xos.Query_DEFAULT, Elements: queryElements}
	onuResponse, err := xosClient.FilterAttWorkflowDriverWhiteListEntry(context.Background(), query)
	onus := onuResponse.GetItems()
	if len(onus) == 0 {
		errorMsg := fmt.Sprintf("Unable to find WhiteListEntry in XOS with SerialNumber %s", ont.SerialNumber)
		return errors.New(errorMsg)
	}
	onu := onus[0]
	log.Printf("DeleteAttWorkflowDriverWhiteListEntry ONU : %v\n", onu)

	id := &xos.ID{Id: onu.GetId()}
	log.Printf("DeleteAttWorkflowDriverWhiteListEntry XOSID:%v\n", id)
	response, err := xosClient.DeleteAttWorkflowDriverWhiteListEntry(context.Background(), id)

	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		return err
	}
	log.Printf("Response is %v\n", response)
	return nil
}

/*
deleteOntTosca - deletes ONT using XOS Tosca Interface
*/
func (chassis *Chassis) deleteOntTosca(ont Ont) {
	ponPort := ont.Parent
	slot := ponPort.Parent
	ontStruct := tosca.NewOntProvision(ont.SerialNumber, slot.Address.IP, ponPort.Number)
	yaml, _ := ontStruct.ToYaml()
	fmt.Println(yaml)

	requestList := fmt.Sprintf("http://%s:%d/delete", chassis.XOSAddress.IP.String(), chassis.XOSAddress.Port)
	client := &http.Client{}
	if settings.GetDummy() {
		log.Printf("yaml:%s\n", yaml)
		log.Println("YAML IS NOT BEING SET TO XOS")
	} else {

		log.Println(requestList)
		log.Println(yaml)
		if settings.GetDummy() {
			return
		}
		req, err := http.NewRequest("POST", requestList, strings.NewReader(yaml))
		req.Header.Add("xos-username", chassis.XOSUser)
		req.Header.Add("xos-password", chassis.XOSPassword)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("ERROR :) %v\n", err)
			// handle error
		}
		log.Printf("Response is %v\n", resp)
	}
	deleteOntStruct := tosca.NewOntDelete(ont.SerialNumber)
	yaml, _ = deleteOntStruct.ToYaml()
	fmt.Println(yaml)
	if settings.GetDummy() {
		log.Printf("yaml:%s\n", yaml)
		log.Println("YAML IS NOT BEING SET TO XOS")
		return
	}
	req, err := http.NewRequest("POST", requestList, strings.NewReader(yaml))
	req.Header.Add("xos-username", chassis.XOSUser)
	req.Header.Add("xos-password", chassis.XOSPassword)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		// handle error
	}
	log.Printf("Response is %v\n", resp)
}
