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
	"log"

	"runtime/debug"

	"gerrit.opencord.org/abstract-olt/api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

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

func main() {
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

	response, err := c.CreateChassis(context.Background(), &api.AddChassisMessage{CLLI: "my cilli", VCoreIP: "192.168.0.1", VCorePort: 9191})
	if err != nil {
		log.Fatalf("Error when calling SayHello: %s", err)
	}
	log.Printf("Response from server: %s", response.GetDeviceID())
	newResponse, err := c.CreateOLTChassis(context.Background(), &api.AddOLTChassisMessage{CLLI: "my cilli", SlotIP: "12.2.2.0", SlotPort: 9191, Hostname: "SlotOne", Type: api.AddOLTChassisMessage_edgecore})
	if err != nil {
		debug.PrintStack()
		log.Fatalf("Error when calling SayHello: %s", err)
	}
	log.Printf("Response from server: %s", newResponse.GetDeviceID())
}
