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
	"strings"

	"gerrit.opencord.org/abstract-olt/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	/* COMMAND FLAGS */
	echo := flag.Bool("e", false, "echo")
	create := flag.Bool("c", false, "create?")
	update := flag.Bool("u", false, "update?")
	addOlt := flag.Bool("s", false, "addOlt?")
	provOnt := flag.Bool("o", false, "provisionOnt?")
	provOntFull := flag.Bool("f", false, "provsionOntFull?")
	preProvOnt := flag.Bool("p", false, "preProvisionOnt?")
	activateSerial := flag.Bool("a", false, "activateSerial?")
	deleteOnt := flag.Bool("d", false, "deleteOnt")
	output := flag.Bool("output", false, "dump output")
	reflow := flag.Bool("reflow", false, "reflow provisioning tosca")
	fullInventory := flag.Bool("full_inventory", false, "pull full inventory json")
	inventory := flag.Bool("inventory", false, "pull json inventory for a specific clli")
	/* END COMMAND FLAGS */

	/* CREATE CHASSIS FLAGS */
	xosUser := flag.String("xos_user", "", "xos_user")
	xosPassword := flag.String("xos_password", "", "xos_password")
	xosAddress := flag.String("xos_address", "", "xos address")
	xosPort := flag.Uint("xos_port", 0, "xos port")
	rack := flag.Uint("rack", 1, "rack number for chassis")
	shelf := flag.Uint("shelf", 1, "shelf number for chassis")
	/* END CREATE CHASSIS FLAGS */

	/* ADD OLT FLAGS */
	oltAddress := flag.String("olt_address", "", "ip address for olt chassis")
	oltPort := flag.Uint("olt_port", 0, "listen port for olt chassis")
	name := flag.String("name", "", "friendly name for olt chassis")
	driver := flag.String("driver", "", "driver to use with olt chassis")
	oltType := flag.String("type", "", "olt chassis type")
	/* END ADD OLT FLAGS */

	/* PROVISION / DELETE ONT FLAGS */
	slot := flag.Uint("slot", 1, "slot number 1-16 to provision ont to")
	port := flag.Uint("port", 1, "port number 1-16 to provision ont to")
	ont := flag.Uint("ont", 1, "ont number 1-64")
	serial := flag.String("serial", "", "serial number of ont")
	/* END PROVISION / DELETE ONT FLAGS */

	/*PROVISION ONT FULL EXTRA FLAGS*/
	stag := flag.Uint("stag", 0, "s-tag for ont")
	ctag := flag.Uint("ctag", 0, "c-tag for ont")
	nasPort := flag.String("nas_port", "", "NasPortID for ont")
	circuitID := flag.String("circuit_id", "", "CircuitID for ont")
	/*END PROVISION ONT FULL EXTRA FLAGS*/

	/*PREPROVISION ONT EXTRA FLAGS*/
	techProfile := flag.String("tech_profile", "", "Tech Profile")
	speedProfile := flag.String("speed_profile", "", "Speed Profile")
	/*END PREPROVISION ONT EXTRA FLAGS*/

	/* ECHO FLAGS */
	message := flag.String("message", "ping", "message to be echoed back")
	/*END ECHO FLAGS*/

	/*GENERIC FLAGS */
	clli := flag.String("clli", "", "clli of abstract chassis")
	useSsl := flag.Bool("ssl", false, "use ssl")
	useAuth := flag.Bool("auth", false, "use auth")
	crtFile := flag.String("cert", "cert/server.crt", "Public cert for server to establish tls session")
	serverAddressPort := flag.String("server", "localhost:7777", "address and port of AbstractOLT server")
	fqdn := flag.String("fqdn", "", "FQDN of the service to match what is in server.crt")
	/*GENERIC FLAGS */

	flag.Parse()

	if *useSsl {
		if *fqdn == "" {
			fqdn = &(strings.Split(*serverAddressPort, ":")[0])
			fmt.Printf("using %s as the FQDN for the AbstractOLT server", *fqdn)
		}
	}

	cmdFlags := []*bool{echo, addOlt, update, create, provOnt, preProvOnt, activateSerial, provOntFull, deleteOnt, output, reflow, fullInventory, inventory}
	cmdCount := 0
	for _, flag := range cmdFlags {
		if *flag {
			cmdCount++
		}
	}
	if cmdCount > 1 {
		fmt.Println("CMD You can only call one method at a time")
		usage()
		return
	}
	if cmdCount < 1 {
		fmt.Println("CMD You didn't specify an operation to perform")
		usage()
		return
	}

	var conn *grpc.ClientConn
	var err error

	// Setup the login/pass
	auth := Authentication{
		Login:    "john",
		Password: "doe",
	}
	if *useSsl && *useAuth {

		creds, err := credentials.NewClientTLSFromFile(*crtFile, *fqdn)
		conn, err = grpc.Dial(*serverAddressPort, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&auth))
		if err != nil {
			log.Fatalf("could not load tls cert: %s", err)
		}
	} else if *useSsl {
		creds, err := credentials.NewClientTLSFromFile("cert/server.crt", *fqdn)
		conn, err = grpc.Dial(*serverAddressPort, grpc.WithTransportCredentials(creds))
		if err != nil {
			log.Fatalf("could not load tls cert: %s", err)
		}
	} else if *useAuth {
		conn, err = grpc.Dial(*serverAddressPort, grpc.WithInsecure(), grpc.WithPerRPCCredentials(&auth))
	} else {
		conn, err = grpc.Dial(*serverAddressPort, grpc.WithInsecure())
	}
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := api.NewAbstractOLTClient(conn)
	if *create {
		createChassis(c, clli, xosUser, xosPassword, xosAddress, xosPort, rack, shelf)
	} else if *update {
		updateXOSUserPassword(c, clli, xosUser, xosPassword)
	} else if *addOlt {
		addOltChassis(c, clli, oltAddress, oltPort, name, driver, oltType)
	} else if *provOnt {
		provisionONT(c, clli, slot, port, ont, serial)
	} else if *provOntFull {
		provisionONTFull(c, clli, slot, port, ont, serial, stag, ctag, nasPort, circuitID)
	} else if *preProvOnt {
		preProvisionOnt(c, clli, slot, port, ont, stag, ctag, nasPort, circuitID, techProfile, speedProfile)
	} else if *activateSerial {
		activateSerialNumber(c, clli, slot, port, ont, serial)
	} else if *echo {
		ping(c, *message)
	} else if *output {
		doOutput(c)
	} else if *deleteOnt {
		deleteONT(c, clli, slot, port, ont, serial)
	} else if *reflow {
		reflowTosca(c)
	} else if *fullInventory {
		getFullInventory(c)
	} else if *inventory {
		getInventory(c, clli)
	}

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
func doOutput(c api.AbstractOLTClient) error {
	response, err := c.Output(context.Background(), &api.OutputMessage{Something: "wtf"})
	if err != nil {
		fmt.Printf("Error when calling Echo: %s", err)
		return err
	}
	log.Printf("Response from server: %v", response.GetSuccess())

	return nil
}

func ping(c api.AbstractOLTClient, message string) error {
	response, err := c.Echo(context.Background(), &api.EchoMessage{Ping: message})
	if err != nil {
		fmt.Printf("Error when calling Echo: %s", err)
		return err
	}
	log.Printf("Response from server: %s", response.GetPong())
	return nil
}

func createChassis(c api.AbstractOLTClient, clli *string, xosUser *string, xosPassword *string, xosAddress *string, xosPort *uint, rack *uint, shelf *uint) error {
	fmt.Println("Calling Create Chassis")
	fmt.Println("clli", *clli)
	fmt.Println("xos_user", *xosUser)
	fmt.Println("xos_password", *xosPassword)
	fmt.Println("xos_address", *xosAddress)
	fmt.Println("xos_port", *xosPort)
	fmt.Println("rack", *rack)
	fmt.Println("shelf", *shelf)
	response, err := c.CreateChassis(context.Background(), &api.AddChassisMessage{CLLI: *clli, XOSUser: *xosUser, XOSPassword: *xosPassword,
		XOSIP: *xosAddress, XOSPort: int32(*xosPort), Rack: int32(*rack), Shelf: int32(*shelf)})
	if err != nil {
		fmt.Printf("Error when calling CreateChassis: %s", err)
		return err
	}
	log.Printf("Response from server: %s", response.GetDeviceID())
	return nil
}
func updateXOSUserPassword(c api.AbstractOLTClient, clli *string, xosUser *string, xosPassword *string) error {
	fmt.Println("Calling Update XOS USER/PASSWORD")
	fmt.Println("clli", *clli)
	fmt.Println("xos_user", *xosUser)
	fmt.Println("xos_password", *xosPassword)
	response, err := c.ChangeXOSUserPassword(context.Background(), &api.ChangeXOSUserPasswordMessage{CLLI: *clli, XOSUser: *xosUser, XOSPassword: *xosPassword})
	if err != nil {
		fmt.Printf("Error when calling UpdateXOSUserPassword: %s", err)
		return err
	}
	log.Printf("Response from server: %t", response.GetSuccess())
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
		driverType = api.AddOLTChassisMessage_openolt

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
}
func preProvisionOnt(c api.AbstractOLTClient, clli *string, slot *uint, port *uint, ont *uint, stag *uint, ctag *uint, nasPort *string, circuitID *string, techProfile *string, speedProfile *string) error {
	fmt.Println("clli", *clli)
	fmt.Println("slot", *slot)
	fmt.Println("port", *port)
	fmt.Println("ont", *ont)
	fmt.Println("stag", *stag)
	fmt.Println("ctag", *ctag)
	fmt.Println("nasPort", *nasPort)
	fmt.Println("circuitID", *circuitID)
	fmt.Println("tech_profile", *techProfile)
	fmt.Println("speed_profile", *speedProfile)
	res, err := c.PreProvisionOnt(context.Background(), &api.PreProvisionOntMessage{CLLI: *clli, SlotNumber: int32(*slot), PortNumber: int32(*port),
		OntNumber: int32(*ont), STag: uint32(*stag), CTag: uint32(*ctag), NasPortID: *nasPort, CircuitID: *circuitID, TechProfile: *techProfile, SpeedProfile: *speedProfile})
	if err != nil {
		debug.PrintStack()
		fmt.Printf("Error when calling ProvsionOnt %s", err)
		return err
	}
	log.Printf("Response from server: %t", res.GetSuccess())
	return nil
}
func activateSerialNumber(c api.AbstractOLTClient, clli *string, slot *uint, port *uint, ont *uint, serial *string) error {
	fmt.Println("clli", *clli)
	fmt.Println("slot", *slot)
	fmt.Println("port", *port)
	fmt.Println("ont", *ont)
	fmt.Println("serial", *serial)
	res, err := c.ActivateSerial(context.Background(), &api.AddOntMessage{CLLI: *clli, SlotNumber: int32(*slot), PortNumber: int32(*port), OntNumber: int32(*ont), SerialNumber: *serial})
	if err != nil {
		debug.PrintStack()
		fmt.Printf("Error when calling ActivateSerial %s", err)
		return err
	}
	log.Printf("Response from server: %t", res.GetSuccess())
	return nil
}
func provisionONTFull(c api.AbstractOLTClient, clli *string, slot *uint, port *uint, ont *uint, serial *string, stag *uint, ctag *uint, nasPort *string, circuitID *string) error {
	fmt.Println("clli", *clli)
	fmt.Println("slot", *slot)
	fmt.Println("port", *port)
	fmt.Println("ont", *ont)
	fmt.Println("serial", *serial)
	fmt.Println("stag", *stag)
	fmt.Println("ctag", *ctag)
	fmt.Println("nasPort", *nasPort)
	fmt.Println("circuitID", *circuitID)
	res, err := c.ProvisionOntFull(context.Background(), &api.AddOntFullMessage{CLLI: *clli, SlotNumber: int32(*slot), PortNumber: int32(*port), OntNumber: int32(*ont), SerialNumber: *serial, STag: uint32(*stag), CTag: uint32(*ctag), NasPortID: *nasPort, CircuitID: *circuitID})
	if err != nil {
		debug.PrintStack()
		fmt.Printf("Error when calling ProvsionOnt %s", err)
		return err
	}
	log.Printf("Response from server: %t", res.GetSuccess())
	return nil
}
func deleteONT(c api.AbstractOLTClient, clli *string, slot *uint, port *uint, ont *uint, serial *string) error {
	fmt.Println("clli", *clli)
	fmt.Println("slot", *slot)
	fmt.Println("port", *port)
	fmt.Println("ont", *ont)
	fmt.Println("serial", *serial)
	res, err := c.DeleteOnt(context.Background(), &api.DeleteOntMessage{CLLI: *clli, SlotNumber: int32(*slot), PortNumber: int32(*port), OntNumber: int32(*ont), SerialNumber: *serial})
	if err != nil {
		debug.PrintStack()
		fmt.Printf("Error when calling ProvsionOnt %s", err)
		return err
	}
	log.Printf("Response from server: %t", res.GetSuccess())
	return nil
}
func reflowTosca(c api.AbstractOLTClient) error {
	res, err := c.Reflow(context.Background(), &api.ReflowMessage{})
	if err != nil {
		debug.PrintStack()
		fmt.Printf("Error when calling Reflow %s", err)
		return err
	}
	log.Printf("Response from server: %t", res.GetSuccess())
	return nil
}
func getFullInventory(c api.AbstractOLTClient) error {
	res, err := c.GetFullInventory(context.Background(), &api.FullInventoryMessage{})
	if err != nil {
		debug.PrintStack()
		fmt.Printf("Error when calling Reflow %s", err)
		return err
	}
	log.Println(res.GetJsonDump())
	return nil
}
func getInventory(c api.AbstractOLTClient, clli *string) error {
	res, err := c.GetInventory(context.Background(), &api.InventoryMessage{Clli: *clli})
	if err != nil {
		debug.PrintStack()
		fmt.Printf("Error when calling Reflow %s", err)
		return err
	}
	log.Println(res.GetJsonDump())
	return nil
}

func usage() {
	var output = `
	Usage ./client -server=[serverAddress:port] -[methodFlag] params
	./client -ssl -fqdn=FQDN_OF_ABSTRACT_OLT_SERVER.CRT -cert PATH_TO_SERVER.CRT -server=[serverAddress:port] -[methodFlag] params : use ssl
	./client -auth -server=[serverAddress:port] -[methodFlag] params : Authenticate session

   methodFlags:
   -e echo # used to test connectivity to server NOOP
      params:
	 -message string to be echoed back from the server
	 e.g. ./client -server=localhost:7777 -e -message MESSAGE_TO_BE_ECHOED

   -c create chassis
      params:
         -clli CLLI_NAME
	 -xos_user XOS_USER
	 -xos_password XOS_PASSWORD
	 -xos_address XOS_TOSCA_IP
	 -xos_port XOS_TOSCA_LISTEN_PORT
	 -rack [optional default 1]
	 -shelf [optional default 1]
	 e.g. ./client -server=localhost:7777 -c -clli MY_CLLI -xos_user foundry -xos_password password -xos_address 192.168.0.1 -xos_port 30007 -rack 1 -shelf 1

   -u update xos user/password
         -clli CLLI_NAME
	 -xos_user XOS_USER
	 -xos_password XOS_PASSWORD
	 e.g. ./client -server=localhost:7777 -u -clli MY_CLLI -xos_user NEW_USER -xos_password NEW_PASSWORD

   -s add physical olt chassis to chassis
      params:
         -clli CLLI_NAME - identifies abstract chassis to assign olt chassis to
	 -olt_address - OLT_CHASSIS_IP_ADDRESS
	 -olt_port - OLT_CHASSIS_LISTEN_PORT
	 -name - OLT_NAME internal human readable name to identify OLT_CHASSIS
	 -driver [openolt,asfvolt16,adtran,tibits] - used to tell XOS which driver should be used to manange chassis
	 -type [edgecore,adtran,tibit] - used to tell AbstractOLT how many ports are available on olt chassis
	 e.g. ./client -server abstractOltHost:7777 -s -clli MY_CLLI -olt_address 192.168.1.100 -olt_port=9191 -name=slot1 -driver=openolt -type=adtran

   -o provision ont - adds ont to whitelist in XOS  on a specific port on a specific olt chassis based on abstract -> phyisical mapping
      params:
	 -clli CLLI_NAME
	 -slot SLOT_NUMBER [1-16]
	 -port OLT_PORT_NUMBER [1-16]
	 -ont ONT_NUMBER [1-64]
	 -serial ONT_SERIAL_NUM
	 e.g. ./client -server=localhost:7777 -o -clli=MY_CLLI -slot=1 -port=1 -ont=22 -serial=aer900jasdf

   -f provision ont full - same as -o above but allows explicit set of s/c vlans , NasPortID and CircuitID
      params:
	 -clli CLLI_NAME
	 -slot SLOT_NUMBER [1-16]
	 -port OLT_PORT_NUMBER [1-16]
	 -ont ONT_NUMBER [1-64]
	 -serial ONT_SERIAL_NUM
	 -stag S_TAG
	 -ctag C_TAG
	 -nas_port NAS_PORT_ID
	 -circuit_id CIRCUIT_ID
	 e.g. ./client -server=localhost:7777 -f -clli=MY_CLLI -slot=1 -port=1 -ont=22 -serial=aer900jasdf -stag=33 -ctag=104 -nas_port="pon 1/1/1/3:1.1" -circuit_id="CLLI 1/1/1/13:1.1"

   -p pre-provision ont - same as -o above but allows explicit set of s/c vlans , NasPortID and CircuitID and NO serial number
      params:
	 -clli CLLI_NAME
	 -slot SLOT_NUMBER [1-16]
	 -port OLT_PORT_NUMBER [1-16]
	 -ont ONT_NUMBER [1-64]
	 -stag S_TAG
	 -ctag C_TAG
	 -nas_port NAS_PORT_ID
	 -circuit_id CIRCUIT_ID
	 -tech_profile TECH_PROFILE
	 -speed_profile SPEED_PROFILE
	 e.g. ./client -server=localhost:7777 -p -clli=MY_CLLI -slot=1 -port=1 -ont=22  -stag=33 -ctag=104 -nas_port="pon 1/1/1/3:1.1" -circuit_id="CLLI 1/1/1/13:1.1 -tech_profile=Business -speed_profile=1GB
   -a activate serial  - adds ont to whitelist in XOS  on a specific port on a specific olt chassis based on abstract -> phyisical mapping - must be preProvisioned
      params:
	 -clli CLLI_NAME
	 -slot SLOT_NUMBER [1-16]
	 -port OLT_PORT_NUMBER [1-16]
	 -ont ONT_NUMBER [1-64]
	 -serial ONT_SERIAL_NUM
	 e.g. ./client -server=localhost:7777 -a -clli=MY_CLLI -slot=1 -port=1 -ont=22 -serial=aer900jasdf

   -d delete ont - removes ont from service
      params:
	 -clli CLLI_NAME
	 -slot SLOT_NUMBER [1-16]
	 -port OLT_PORT_NUMBER [1-16]
	 -ont ONT_NUMBER [1-64]
	 -serial ONT_SERIAL_NUM
	 e.g. ./client -server=localhost:7777 -d -clli=MY_CLLI -slot=1 -port=1 -ont=22 -serial=aer900jasdf

    -output (TEMPORARY) causes AbstractOLT to serialize all chassis to JSON file in $WorkingDirectory/backups
         e.g. ./client -server=localhost:7777 -output

    -reflow causes tosca to be repushed to xos
	e.g. ./client -server=localhost:7777 -reflow

    -inventory - returns a json document that describes currently provisioned equipment for a specific clli
      params:
	 -clli CLLI_NAME
	 e.g. ./client -inventory -clli=ATLEDGEVOLT1

    -full_inventory - returns a json document that describes all currently provisioned pods
         e.g. ./client -full_inventory

	 `

	fmt.Println(output)
}
