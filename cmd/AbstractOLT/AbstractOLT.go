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
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"gerrit.opencord.org/abstract-olt/api"
	"gerrit.opencord.org/abstract-olt/internal/pkg/settings"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
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
	if err != nil {
		log.Printf("startGRPCServer failed to start with %v\n", err)
		return fmt.Errorf("failed to listen: %v", err)
	}

	// create a server instance
	s := api.Server{}

	// Create the TLS credentials
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("could not load TLS keys: %s", err)
	}

	// Create an array of gRPC options with the credentials
	var opts []grpc.ServerOption
	if *useSsl && *useAuthentication {
		opts = []grpc.ServerOption{grpc.Creds(creds),
			grpc.UnaryInterceptor(unaryInterceptor)}
	} else if *useAuthentication {
		opts = []grpc.ServerOption{grpc.UnaryInterceptor(unaryInterceptor)}
	} else if *useSsl {
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

	if *useAuthentication {
		mux = runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(credMatcher))
	} else {
		mux = runtime.NewServeMux()
	}
	creds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		return fmt.Errorf("could not load TLS certificate: %s", err)
	}

	// Setup the client gRPC options
	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	err = api.RegisterAbstractOLTHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
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

	flag.Parse()

	if *help || *h {
		var usage = `./AbstractOLT -d [default false] : Runs in Debug mode
Params:
      -s [default false] -cert_dir [default $WORKING_DIR/cert]  DIR : Runs in SSL mode with server.crt and server.key found in  DIR
      -a [default false] : Run in Authentication mode currently very basic
      -listenAddress IP_ADDRESS [default localhost] -grpc_port [default 7777] PORT1 -rest_port [default 7778] PORT2: Listen for grpc on IP_ADDRESS:PORT1 and rest on IP_ADDRESS:PORT2
      -log_file [default $WORKING_DIR/AbstractOLT.log] LOG_FILE
      -h(elp) print this usage
`
		fmt.Println(usage)
		return
	}
	settings.SetDebug(*debugPtr)
	fmt.Println("Startup Params: debug:", *debugPtr, " Authentication:", *useAuthentication, " SSL:", *useSsl, "Cert Directory", *certDirectory,
		"ListenAddress:", *listenAddress, " grpc port:", *grpcPort, " rest port:", *restPort, "Logging to ", *logFile)

	file, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", file, ":", err)
	}
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
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
			log.Fatalf("failed to start gRPC server: %s", err)
		}
	}()

	// fire the REST server in a goroutine
	go func() {
		err := startRESTServer(restAddress, grpcAddress, certFile)
		if err != nil {
			log.Fatalf("failed to start gRPC server: %s", err)
		}
	}()

	// infinite loop
	log.Printf("Entering infinite loop")
	select {}
	//TODO publish periodic stats etc
	fmt.Println("AbstractOLT")
}
