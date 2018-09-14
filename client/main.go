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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime/debug"

	"gerrit.opencord.org/abstract-olt/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	create := flag.Bool("c", false, "create?")
	addOlt := flag.Bool("s", false, "addOlt?")
	provOnt := flag.Bool("o", false, "provisionOnt?")
	clli := flag.String("clli", "", "clli of abstract chassis")
	xosAddress := flag.String("xos_address", "", "xos address")
	xosPort := flag.Uint("xos_port", 0, "xos port")
	rack := flag.Uint("rack", 1, "rack number for chassis")
	shelf := flag.Uint("shelf", 1, "shelf number for chassis")
	oltAddress := flag.String("olt_address", "", "ip address for olt chassis")
	oltPort := flag.Uint("olt_port", 0, "listen port for olt chassis")
	name := flag.String("name", "", "friendly name for olt chassis")
	driver := flag.String("driver", "", "driver to use with olt chassis")
	oltType := flag.String("type", "", "olt chassis type")
	slot := flag.Uint("slot", 1, "slot number 1-16 to provision ont to")
	port := flag.Uint("port", 1, "port number 1-16 to provision ont to")
	ont := flag.Uint("ont", 1, "ont number 1-64")
	serial := flag.String("serial", "", "serial number of ont")

	flag.Parse()

	if (*create && *addOlt) || (*create && *provOnt) || (*addOlt && *provOnt) {
		fmt.Println("You can only call one method at a time")
		usage()
		return
	}
	if !(*create || *provOnt || *addOlt) {
		fmt.Println("You didn't specify an operation to perform")
		usage()
		return
	}
	var conn *grpc.ClientConn
	creds, err := credentials.NewClientTLSFromFile("cert/server.crt", "AbstractOLT.dev.atl.foundry.att.com")
	if err != nil {
		log.Fatalf("could not load tls cert: %s", err)
	}
	// Setup the login/pass
	auth := Authentication{
		Login:    "john",
		Password: "doe",
	}
	conn, err = grpc.Dial(":7777", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&auth))
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := api.NewAbstractOLTClient(conn)
	if *create {
		createChassis(c, clli, xosAddress, xosPort, rack, shelf)
	} else if *addOlt {
		addOltChassis(c, clli, oltAddress, oltPort, name, driver, oltType)
	} else if *provOnt {
		provisionONT(c, clli, slot, port, ont, serial)
	} else {
	}

	fmt.Println("TODO - Do something")
}

// Authentication holds the login/password
type Authentication struct {
	Login    string
	Password string
}

// GetRequestMetadata gets the current request metadata
func (a *Authentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"login":    a.Login,
		"password": a.Password,
	}, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security
func (a *Authentication) RequireTransportSecurity() bool {
	return true
}

func createChassis(c api.AbstractOLTClient, clli *string, xosAddress *string, xosPort *uint, rack *uint, shelf *uint) error {
	fmt.Println("Calling Create Chassis")
	fmt.Println("clli", *clli)
	fmt.Println("xos_address", *xosAddress)
	fmt.Println("xos_port", *xosPort)
	fmt.Println("rack", *rack)
	fmt.Println("shelf", *shelf)
	response, err := c.CreateChassis(context.Background(), &api.AddChassisMessage{CLLI: *clli, VCoreIP: *xosAddress, VCorePort: int32(*xosPort), Rack: int32(*rack), Shelf: int32(*shelf)})
	if err != nil {
		fmt.Printf("Error when calling CreateChassis: %s", err)
		return err
	}
	log.Printf("Response from server: %s", response.GetDeviceID())
	return nil
}
func addOltChassis(c api.AbstractOLTClient, clli *string, oltAddress *string, oltPort *uint, name *string, driver *string, oltType *string) error {
	fmt.Println("clli", *clli)
	fmt.Println("olt_address", *oltAddress)
	fmt.Println("olt_port", *oltPort)
	fmt.Println("name", *name)
	fmt.Println("driver", *driver)
	fmt.Println("type", *oltType)
	var driverType api.AddOLTChassisMessage_OltDriver
	var chassisType api.AddOLTChassisMessage_OltType
	switch *oltType {
	case "edgecore":
		chassisType = api.AddOLTChassisMessage_edgecore
	}
	switch *driver {
	case "openolt":
		driverType = api.AddOLTChassisMessage_openoltDriver

	}

	res, err := c.CreateOLTChassis(context.Background(), &api.AddOLTChassisMessage{CLLI: *clli, SlotIP: *oltAddress, SlotPort: uint32(*oltPort), Hostname: *name, Type: chassisType, Driver: driverType})
	if err != nil {
		debug.PrintStack()
		fmt.Printf("Error when calling CreateOLTChassis: %s", err)
		return err
	}
	log.Printf("Response from server: %s", res.GetDeviceID())
	return nil
}
func provisionONT(c api.AbstractOLTClient, clli *string, slot *uint, port *uint, ont *uint, serial *string) error {
	fmt.Println("clli", *clli)
	fmt.Println("slot", *slot)
	fmt.Println("port", *port)
	fmt.Println("ont", *ont)
	fmt.Println("serial", *serial)
	res, err := c.ProvisionOnt(context.Background(), &api.AddOntMessage{CLLI: *clli, SlotNumber: int32(*slot), PortNumber: int32(*port), OntNumber: int32(*ont), SerialNumber: *serial})
	if err != nil {
		debug.PrintStack()
		fmt.Printf("Error when calling ProvsionOnt %s", err)
		return err
	}
	log.Printf("Response from server: %t", res.GetSuccess())
	return nil

	return nil
}
func usage() {
	var output = `
   Usage ./client -[methodFlag] params
   methFlags:
   -c create chassis
      params:
         -clli CLLI_NAME
	 -xos_address XOS_TOSCA_IP
	 -xos_port XOS_TOSCA_LISTEN_PORT
	 -rack [optional default 1]
	 -shelf [optional default 1]
   e.g. ./client -c -clli MY_CLLI -xos_address 192.168.0.1 -xos_port 30007 -rack 1 -shelf 1
   -s add physical olt chassis to chassis
      params:
         -clli CLLI_NAME - identifies abstract chassis to assign olt chassis to
	 -olt_address - OLT_CHASSIS_IP_ADDRESS
	 -olt_port - OLT_CHASSIS_LISTEN_PORT
	 -name - OLT_NAME internal human readable name to identify OLT_CHASSIS
	 -driver [openolt,asfvolt16,adtran,tibits] - used to tell XOS which driver should be used to manange chassis
	 -type [edgecore,adtran,tibit] - used to tell AbstractOLT how many ports are available on olt chassis
   e.g. ./client -s -clli MY_CLLI -olt_address 192.168.1.100 -olt_port=9191 -name=slot1 -driver=openolt -type=adtran
   -o provision ont - adds ont to whitelist in XOS  on a specific port on a specific olt chassis based on abstract -> phyisical mapping
      params:
	 -clli CLLI_NAME
	 -slot SLOT_NUMBER [1-16]
	 -port OLT_PORT_NUMBER [1-16]
	 -ont ONT_NUMBER [1-64]
	 -serial ONT_SERIAL_NUM
	e.g. ./client -o -clli=MY_CLLI -slot=1 -port=1 -ont=22 -serial=aer900jasdf `

	fmt.Println(output)
}
