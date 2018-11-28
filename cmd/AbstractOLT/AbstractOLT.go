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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"gerrit.opencord.org/abstract-olt/api"
	"gerrit.opencord.org/abstract-olt/internal/pkg/impl"
	"gerrit.opencord.org/abstract-olt/internal/pkg/settings"
	"gerrit.opencord.org/abstract-olt/models"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/x/bsonx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// private type for Context keys
type contextKey int

var useSsl *bool
var useAuthentication *bool
var certDirectory *string

const (
	clientIDKey contextKey = iota
)

/*
GetLogger - returns the logger
*/
func credMatcher(headerName string) (mdName string, ok bool) {
	if headerName == "Login" || headerName == "Password" {
		return headerName, true
	}
	return "", false
}

// authenticateAgent check the client credentials
func authenticateClient(ctx context.Context, s *api.Server) (string, error) {
	//TODO if we decide to handle Authentication with AbstractOLT this will need to be bound to an authentication service
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		clientLogin := strings.Join(md["login"], "")
		clientPassword := strings.Join(md["password"], "")

		if clientLogin != "john" {
			return "", fmt.Errorf("unknown user %s", clientLogin)
		}
		if clientPassword != "doe" {
			return "", fmt.Errorf("bad password %s", clientPassword)
		}

		log.Printf("authenticated client: %s", clientLogin)
		return "42", nil
	}
	return "", fmt.Errorf("missing credentials")
}

// unaryInterceptor call authenticateClient with current context
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	s, ok := info.Server.(*api.Server)
	if !ok {
		return nil, fmt.Errorf("unable to cast server")
	}
	clientID, err := authenticateClient(ctx, s)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, clientIDKey, clientID)
	return handler(ctx, req)
}

func startGRPCServer(address, certFile, keyFile string) error {
	if settings.GetDebug() {
		log.Printf("startGRPCServer(LisenAddress:%s,CertFile:%s,KeyFile:%s\n", address, certFile, keyFile)
	}
	// create a listener on TCP port
	lis, err := net.Listen("tcp", address)

	// create a server instance
	s := api.Server{}

	// Create the TLS credentials

	// Create an array of gRPC options with the credentials
	var opts []grpc.ServerOption
	creds, err := credentials.NewClientTLSFromFile(certFile, "")
	if *useSsl && *useAuthentication {
		if err != nil {
			return fmt.Errorf("could not load TLS keys: %s", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds),
			grpc.UnaryInterceptor(unaryInterceptor)}
	} else if *useAuthentication {
		opts = []grpc.ServerOption{grpc.UnaryInterceptor(unaryInterceptor)}
	} else if *useSsl {
		if err != nil {
			return fmt.Errorf("could not load TLS keys: %s", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	} else {
		opts = []grpc.ServerOption{}
	}

	// create a gRPC server object
	grpcServer := grpc.NewServer(opts...)

	// attach the Ping service to the server
	api.RegisterAbstractOLTServer(grpcServer, &s)

	// start the server
	log.Printf("starting HTTP/2 gRPC server on %s", address)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %s", err)
	}

	return nil
}
func startRESTServer(address, grpcAddress, certFile string) error {
	if settings.GetDebug() {
		log.Printf("startRESTServer(Address:%s, GRPCAddress:%s,Cert File:%s\n", address, grpcAddress, certFile)
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var mux *runtime.ServeMux

	var opts []grpc.DialOption
	if *useAuthentication {
		mux = runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(credMatcher))
	} else {
		mux = runtime.NewServeMux()
	}
	if *useSsl {
		creds, err := credentials.NewClientTLSFromFile(certFile, "")
		opts = []grpc.DialOption{grpc.WithTransportCredentials(creds)}
		if err != nil {
			return fmt.Errorf("could not load TLS certificate: %s", err)
		}
	} else {
		opts = []grpc.DialOption{grpc.WithInsecure()}
	}
	// Setup the client gRPC options
	err := api.RegisterAbstractOLTHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		return fmt.Errorf("could not register service Ping: %s", err)
	}

	log.Printf("starting HTTP/1.1 REST server on %s", address)
	http.ListenAndServe(address, mux)

	return nil
}
func main() {
	debugPtr := flag.Bool("d", false, "Log Level Debug")
	useAuthentication = flag.Bool("a", false, "Use Authentication")
	useSsl = flag.Bool("s", false, "Use SSL")
	certDirectory = flag.String("cert_dir", "cert", "Directory where key files exist")
	listenAddress := flag.String("listenAddress", "localhost", "IP Address to listen on")
	grpcPort := flag.String("grpc_port", "7777", "Port to listen for GRPC")
	restPort := flag.String("rest_port", "7778", "Port to listen for Rest Server")
	logFile := flag.String("log_file", "AbstractOLT.log", "Name of the LogFile to write to")
	h := flag.Bool("h", false, "Show usage")
	help := flag.Bool("help", false, "Show usage")
	dummy := flag.Bool("dummy", false, "Run in dummy mode where YAML is not sent to XOS")
	grpc := flag.Bool("grpc", false, "Use XOS GRPC interface instead of TOSCA")

	useMongo := flag.Bool("useMongo", false, "use mongo db for backup/restore")
	mongodb := flag.String("mongodb", "mongodb://foundry:foundry@localhost:27017", "connect string for mongodb backup/restore")

	flag.Parse()
	settings.SetDummy(*dummy)

	if *help || *h {
		var usage = `./AbstractOLT -d [default false] : Runs in Debug mode
Params:
      -s [default false] -cert_dir [default $WORKING_DIR/cert]  DIR : Runs in SSL mode with server.crt and server.key found in  DIR
      -a [default false] : Run in Authentication mode currently very basic
      -listenAddress IP_ADDRESS [default localhost] -grpc_port [default 7777] PORT1 -rest_port [default 7778] PORT2: Listen for grpc on IP_ADDRESS:PORT1 and rest on IP_ADDRESS:PORT2
      -log_file [default $WORKING_DIR/AbstractOLT.log] LOG_FILE
      -mongo [default false] use mongodb for backup restore
      -mongodb [default mongodb://foundry:foundry@localhost:27017] connect string for mongodb - Required if mongo == true
      -grpc [default false] tell AbstractOLT to use XOS GRPC interface instead of TOSCA
      -h(elp) print this usage

`
		fmt.Println(usage)
		return
	}
	settings.SetDebug(*debugPtr)
	settings.SetMongo(*useMongo)
	settings.SetMongodb(*mongodb)
	settings.SetGrpc(*grpc)
	fmt.Println("Startup Params: debug:", *debugPtr, " Authentication:", *useAuthentication, " SSL:", *useSsl, "Cert Directory", *certDirectory,
		"ListenAddress:", *listenAddress, " grpc port:", *grpcPort, " rest port:", *restPort, "Logging to ", *logFile, "Use XOS GRPC ", *grpc)

	file, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", file, ":", err)
	}
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	if *dummy {
		fmt.Println("RUNNING IN DUMMY MODE NO YAML WILL BE SENT TO XOS")
		log.Println("RUNNING IN DUMMY MODE NO YAML WILL BE SENT TO XOS")
	}
	log.Printf("Setting Debug to %t\n", settings.GetDebug())
	if settings.GetDebug() {
		log.Println("Startup Params: debug:", *debugPtr, " Authentication:", *useAuthentication, " SSL:", *useSsl, "Cert Directory", *certDirectory,
			"ListenAddress:", *listenAddress, " grpc port:", *grpcPort, " rest port:", *restPort, "Logging to ", *logFile)
	}

	grpcAddress := fmt.Sprintf("%s:%s", *listenAddress, *grpcPort)
	restAddress := fmt.Sprintf("%s:%s", *listenAddress, *restPort)

	certFile := fmt.Sprintf("%s/server.crt", *certDirectory)
	keyFile := fmt.Sprintf("%s/server.key", *certDirectory)

	// fire the gRPC server in a goroutine
	go func() {
		err := startGRPCServer(grpcAddress, certFile, keyFile)
		if err != nil {
			log.Printf("failed to start gRPC server: %s", err)
		}
	}()

	// fire the REST server in a goroutine
	go func() {
		err := startRESTServer(restAddress, grpcAddress, certFile)
		if err != nil {
			log.Printf("failed to start REST server: %s", err)
		}
	}()

	// infinite loop
	if *useMongo {
		clientOptions := options.Client()
		creds := options.Credential{AuthMechanism: "SCRAM-SHA-256", AuthSource: "AbstractOLT", Username: "seba", Password: "seba"}
		clientOptions.SetAuth(creds)

		client, err := mongo.NewClientWithOptions(*mongodb, clientOptions)

		client.Connect(context.Background())
		fmt.Println(client)
		defer client.Disconnect(context.Background())
		if err != nil {
			log.Fatalf("unable to connect to mongodb with %v\n", err)
		}
		collection := client.Database("AbstractOLT").Collection("backups")
		cur, err := collection.Find(context.Background(), nil)
		if err != nil {
			log.Fatalf("Unable to connect to collection with %v\n", err)
		}
		defer cur.Close(context.Background())
		for cur.Next(context.Background()) {
			doc := bsonx.Doc{}
			err := cur.Decode(&doc)
			if err != nil {
				log.Fatal(err)
			}
			clli := doc.LookupElement("_id").Value
			body := doc.LookupElement("body").Value
			_, bodyBin := (body).Binary()

			chassisHolder := models.ChassisHolder{}
			err = chassisHolder.Deserialize(bodyBin)
			if err != nil {
				log.Printf("Deserialize threw an error for clli %s %v\n", (clli).StringValue(), err)
			} else {
				chassisMap := models.GetChassisMap()
				(*chassisMap)[(clli).StringValue()] = &chassisHolder

			}
		}
	} else {
		files, err := ioutil.ReadDir("backup")
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			chassisHolder := models.ChassisHolder{}
			if file.Name() != "BackupPlaceHolder" {
				fileName := fmt.Sprintf("backup/%s", file.Name())
				json, _ := ioutil.ReadFile(fileName)
				err := chassisHolder.Deserialize([]byte(json))
				if err != nil {
					fmt.Printf("Deserialize threw an error %v\n", err)
				}
				chassisMap := models.GetChassisMap()
				(*chassisMap)[file.Name()] = &chassisHolder
			} else {
				fmt.Println("Ignoring BackupPlaceHolder")
			}
		}
	}

	log.Printf("Entering infinite loop")
	var ticker = time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ticker.C:
			impl.DoOutput()
		}
	}

	//TODO publish periodic stats etc
	fmt.Println("AbstractOLT")
}
