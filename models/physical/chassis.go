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
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"gerrit.opencord.org/abstract-olt/internal/pkg/settings"
	"gerrit.opencord.org/abstract-olt/models/tosca"
)

/*
Chassis is a model that takes up to 16 discreet OLT chassis as if it is a 16 slot OLT chassis
*/
type Chassis struct {
	CLLI         string
	VCoreAddress net.TCPAddr
	Dataswitch   DataSwitch
	Linecards    []OLT
	Rack         int
	Shelf        int
}
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
func (chassis *Chassis) AddOLTChassis(olt OLT) {
	chassis.Linecards = append(chassis.Linecards, olt)
	//TODO - api call to add olt i.e. preprovision_olt
	fmt.Printf("chassis.AddOLTChassis(%s)\n", olt.GetHostname())
	//S>103 func NewOltProvision(name string, deviceType string, host string, port int) OltProvsion {
	ipString := olt.GetAddress().IP.String()
	webServerPort := olt.GetAddress().Port
	oltStruct := tosca.NewOltProvision(chassis.CLLI, olt.GetHostname(), "openolt", ipString, webServerPort)
	yaml, _ := oltStruct.ToYaml()
	if settings.GetDummy() {
		log.Printf("yaml:%s\n", yaml)
		log.Println("YAML IS NOT BEING SET TO XOS")
		return
	}
	client := &http.Client{}
	requestList := fmt.Sprintf("http://%s:%d/run", chassis.VCoreAddress.IP.String(), chassis.VCoreAddress.Port)
	req, err := http.NewRequest("POST", requestList, strings.NewReader(yaml))
	req.Header.Add("xos-username", "admin@opencord.org")
	req.Header.Add("xos-password", "letmein")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("ERROR :) %v\n", err)
		// handle error
	}
	fmt.Printf("Response is %v\n", resp)

}
func (chassis *Chassis) provisionONT(ont Ont) {
	//TODO - api call to provison s/c vlans and ont serial number etc
	fmt.Printf("chassis.provisionONT(%s,SVlan:%d,CVlan:%d)\n", ont.SerialNumber, ont.Svlan, ont.Cvlan)
	ponPort := ont.Parent
	slot := ponPort.Parent

	//func NewOntProvision(serialNumber string, oltIP net.IP, ponPortNumber int) OntProvision {
	ontStruct := tosca.NewOntProvision(ont.SerialNumber, slot.Address.IP, ponPort.Number)
	yaml, _ := ontStruct.ToYaml()

	if settings.GetDummy() {
		log.Printf("yaml:%s\n", yaml)
		log.Println("YAML IS NOT BEING SET TO XOS")
		return
	}
	client := &http.Client{}
	requestList := fmt.Sprintf("http://%s:%d/run", chassis.VCoreAddress.IP.String(), chassis.VCoreAddress.Port)
	req, err := http.NewRequest("POST", requestList, strings.NewReader(yaml))
	req.Header.Add("xos-username", "admin@opencord.org")
	req.Header.Add("xos-password", "letmein")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("ERROR :) %v\n", err)
		// handle error
	}
	fmt.Printf("Response is %v\n", resp)
	rgName := fmt.Sprintf("%s_%d_%d_%d_RG", chassis.CLLI, slot.Number, ponPort.Number, ont.Number)
	subStruct := tosca.NewSubscriberProvision(rgName, ont.Cvlan, ont.Svlan, ont.SerialNumber, ont.NasPortID, ont.CircuitID, chassis.CLLI)
	yaml, _ = subStruct.ToYaml()
	fmt.Printf("yaml:%s\n", yaml)
	req, err = http.NewRequest("POST", requestList, strings.NewReader(yaml))
	req.Header.Add("xos-username", "admin@opencord.org")
	req.Header.Add("xos-password", "letmein")
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("ERROR :) %v\n", err)
		// handle error
	}
}
func (chassis *Chassis) deleteONT(ont Ont) {
	//TODO - api call to provison s/c vlans and ont serial number etc
	fmt.Printf("chassis.deleteONT(%s,SVlan:%d,CVlan:%d)\n", ont.SerialNumber, ont.Svlan, ont.Cvlan)
}
func (chassis *Chassis) ActivateSlot(slotNumber int) error {
	// AT&T backend systems start at 1 and not 0 :P
	if chassis.Linecards[slotNumber-1] == nil {
		return &UnprovisionedSlotError{CLLI: chassis.CLLI, SlotNumber: slotNumber}
	}
	return chassis.Linecards[slotNumber-1].activate()
}
