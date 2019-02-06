#
#   Copyright 2017 the original author or authors.
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#        http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#

SERVER_OUT	 := "bin/AbstractOLT"
CLIENT_OUT	 := "bin/client"
API_OUT		 := "api/abstract_olt_api.pb.go"
API_REST_OUT     := "api/abstract_olt_api.pb.gw.go"
SWAGGER_OUT      := "api/abstract_olt_api.swagger.json"
PKG	         := "gerrit.opencord.org/abstract-olt"
SERVER_PKG_BUILD := "${PKG}/cmd/AbstractOLT"
CLIENT_PKG_BUILD := "${PKG}/client"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
DOCKERTAG ?= "latest"
SEBA_PROTO_PATH := contrib/xos
SEBA_SCHEMA_PATH := contrib/schema
SEBA_PROTO_FILES := $(wildcard contrib/xos/*.proto)
SEBA_PROTO_GO_FILES := $(foreach f,$(SEBA_PROTO_FILES),$(subst .proto,.pb.go,$(f)))
SEBA_PROTO_DESC_FILES := $(foreach f,$(SEBA_PROTO_FILES),$(subst .proto,.desc,$(f)))


.PHONY: all api server client test docker

all: server client

api/abstract_olt_api.pb.go:
	@protoc -I api/ \
	-I${GOPATH}/src \
	-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	--go_out=plugins=grpc:api \
	api/abstract_olt_api.proto

api/abstract_olt_api.pb.gw.go :
	  @protoc -I api/ \
	-I${GOPATH}/src \
	-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	--grpc-gateway_out=logtostderr=true:api \
	 api/abstract_olt_api.proto

seba-api:  schema/schema.pb.go $(SEBA_PROTO_GO_FILES)

%.pb.go: %.proto
	@protoc -I ${SEBA_PROTO_PATH} \
                -I ${SEBA_SCHEMA_PATH} \
                --go_out=plugins=grpc:${SEBA_PROTO_PATH} \
                 -I${GOPATH}/src \
                 -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
                 $<

schema/schema.pb.go:
	@protoc -I ${SEBA_PROTO_PATH} \
                -I ${SEBA_SCHEMA_PATH} \
                --go_out=plugins=grpc:${SEBA_SCHEMA_PATH} \
                -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
                 --descriptor_set_out=${SEBA_SCHEMA_PATH}/schema.desc \
                 --include_imports \
                 --include_source_info \
                 ${SEBA_SCHEMA_PATH}/schema.proto



swagger:
	@protoc -I api/ \
  -I${GOPATH}/src \
  -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --swagger_out=logtostderr=true:api \
  api/abstract_olt_api.proto

api: api/abstract_olt_api.pb.go api/abstract_olt_api.pb.gw.go swagger

dep: ## Get the dependencies
	@go get -v -d ./...

server: api seba-api dep ## Build the binary file for server
	@go build -i -v -o $(SERVER_OUT) $(SERVER_PKG_BUILD)

client:  api dep## Build the binary file for client
	@go build -i -v -o $(CLIENT_OUT) $(CLIENT_PKG_BUILD)

clean: ## Remove previous builds
	@rm $(SERVER_OUT) $(CLIENT_OUT) $(API_OUT) $(API_REST_OUT) $(SWAGGER_OUT)
	@rm contrib/xos/*.go
	@rm contrib/schema/*.go


test: all
	@go test ./...
	@go test ./... -cover

docker: ## build docker image
	@docker build -t opencord/abstract-olt:${DOCKERTAG} .

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
