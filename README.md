# AbstractOLT
AbstractOLT provides a mapping service from AT&T's, and potentially other's, provisioning platform to SEBA/Voltha.
It presents a facade of the hardware that resembles a 16 slot / with 16 PON ports each.
Internally it maps PON ports on the abstractOLT to PON ports on physical OLT chasssis.


## Dependencies
```go
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
go get -u github.com/golang/protobuf/protoc-gen-go
```
## Retrieving code
```go
go get gerrit.opencord.org/abstract-olt/cmd/AbstractOLT
```
## Makefile
```sh
make clean # clears all generated and compiler output.
make dep # pulls any go dependencies.
make api # generates grpc code and swagger files
make server # compiles AbstractOLT into bin directory
make client # compiles a test client which exercises some of the server api functions
make all # builds everything
make docker # builds a docker container with the AbstractOLT and client in /app
```
### Certificates
Before running you will need to create a cert directory under bin and generate ssl certificates
```sh
cd bin
mkdir cert
openssl genrsa -out cert/server.key 204
openssl req -new -x509 -sha256 -key cert/server.key -out cert/server.crt -days 3650
openssl req -new -sha256 -key cert/server.key -out cert/server.csr
openssl x509 -req -sha256 -in cert/server.csr -signkey cert/server.key -out cert/server.crt -days 3650
```

## Running
```sh
cd bin
./AbstractOLT # or
./AbstractOLT -d # for debug logging
```
This will create/append an AbstractOLT.log file in the current directory that will contain some runtime information





