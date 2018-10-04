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
	CLLI        string
	XOSAddress  net.TCPAddr
	XOSUser     string
	XOSPassword string
	Linecards   []SimpleOLT
	Rack        int
	Shelf       int
}
type UnprovisionedSlotError struct {
	CLLI       string
	SlotNumber int
}

func (c *Chassis) Output() {
	for _, olt := range c.Linecards {
		olt.Output()
	}
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
	//TODO - api call to add olt i.e. preprovision_olt
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
	requestList := fmt.Sprintf("http://%s:%d/run", chassis.XOSAddress.IP.String(), chassis.XOSAddress.Port)
	req, err := http.NewRequest("POST", requestList, strings.NewReader(yaml))
	req.Header.Add("xos-username", chassis.XOSUser)
	req.Header.Add("xos-password", chassis.XOSPassword)
	resp, err := client.Do(req)
	if err != nil {
		//TODO
		// handle error
	}
	log.Printf("Server response was %v\n", resp)

}
func (chassis *Chassis) provisionONT(ont Ont) {
	//TODO - api call to provison s/c vlans and ont serial number etc
	log.Printf("chassis.provisionONT(%s,SVlan:%d,CVlan:%d)\n", ont.SerialNumber, ont.Svlan, ont.Cvlan)
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
	requestList := fmt.Sprintf("http://%s:%d/run", chassis.XOSAddress.IP.String(), chassis.XOSAddress.Port)
	req, err := http.NewRequest("POST", requestList, strings.NewReader(yaml))
	req.Header.Add("xos-username", chassis.XOSUser)
	req.Header.Add("xos-password", chassis.XOSPassword)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		// handle error
	}
	log.Printf("Response is %v\n", resp)
	rgName := fmt.Sprintf("%s_%d_%d_%d_RG", chassis.CLLI, slot.Number, ponPort.Number, ont.Number)
	subStruct := tosca.NewSubscriberProvision(rgName, ont.Cvlan, ont.Svlan, ont.SerialNumber, ont.NasPortID, ont.CircuitID, chassis.CLLI)
	yaml, _ = subStruct.ToYaml()
	log.Printf("yaml:%s\n", yaml)
	req, err = http.NewRequest("POST", requestList, strings.NewReader(yaml))
	req.Header.Add("xos-username", chassis.XOSUser)
	req.Header.Add("xos-password", chassis.XOSPassword)
	resp, err = client.Do(req)
	if err != nil {
		log.Printf("ERROR :) %v\n", err)
		// handle error
	}
}
func (chassis *Chassis) deleteONT(ont Ont) {
	//TODO - api call to provison s/c vlans and ont serial number etc
	//TODO - api call to provison s/c vlans and ont serial number etc
	log.Printf("chassis.deleteONT(%s,SVlan:%d,CVlan:%d)\n", ont.SerialNumber, ont.Svlan, ont.Cvlan)
	ponPort := ont.Parent
	slot := ponPort.Parent

	//func NewOntProvision(serialNumber string, oltIP net.IP, ponPortNumber int) OntProvision {
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
	} else {
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
}
