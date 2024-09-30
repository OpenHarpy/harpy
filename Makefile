.PHONY: all clean

# Compiler 
GO := GODEBUG=gctrace=1 go
PYTHON := poetry run python
GOFLAGS := -buildmode=c-shared
GO_PROJECT_CLIENT_ENGING := $(shell pwd)/client-engine
GO_PROJECT_REMOTE_RUNNER := $(shell pwd)/remote-runner
GO_PROJECT_RESOURCE_MANAGER := $(shell pwd)/resource-manager

PYTHON_PROJECT_DIR := $(shell pwd)/py-client/sdk
PYTHON_PROJECT_ROOT := $(shell pwd)/py-client

DISTRIBUTION_DIR := $(shell pwd)/dist

# Protobuf files
PROTO_DIR := $(shell pwd)/grpc_protos
PYTHON_PROTO := $(PYTHON_PROJECT_DIR)/grpc_ce_protocol
GRPC_PYTHON_PLUGIN=$(shell $(PYTHON) -m grpc_tools.protoc --plugin)
NODE_TESTING_DIR := $(GO_PROJECT_REMOTE_RUNNER)/py_src/venv

# Abstract protoc 
define protoc_command
    protoc -I=$(PROTO_DIR) \
        --go_out=$(1) \
        --go_opt=paths=source_relative \
        --go_opt=Mgrpc_ce_protocol/ceprotocol.proto=$(3)/grpc_ce_protocol \
        --go_opt=Mgrpc_node_protocol/nodeprotocol.proto=$(3)/grpc_node_protocol \
        --go_opt=Mgrpc_resource_alloc_procotol/resourceallocprotocol.proto=$(3)/grpc_resource_alloc_procotol \
        --go-grpc_out=$(1) \
        --go-grpc_opt=paths=source_relative \
        --go-grpc_opt=Mgrpc_ce_protocol/ceprotocol.proto=$(3)/grpc_ce_protocol \
        --go-grpc_opt=Mgrpc_node_protocol/nodeprotocol.proto=$(3)/grpc_node_protocol \
        --go-grpc_opt=Mgrpc_resource_alloc_procotol/resourceallocprotocol.proto=$(3)/grpc_resource_alloc_procotol \
        $(2)
endef

# Build the protobuf files
build-proto-go:
	$(call protoc_command,$(GO_PROJECT_CLIENT_ENGING),$(PROTO_DIR)/*/*.proto,client-engine)
	$(call protoc_command,$(GO_PROJECT_REMOTE_RUNNER),$(PROTO_DIR)/grpc_node_protocol/*.proto,remote-runner)
	$(call protoc_command,$(GO_PROJECT_REMOTE_RUNNER),$(PROTO_DIR)/grpc_resource_alloc_procotol/*.proto,remote-runner)
	$(call protoc_command,$(GO_PROJECT_RESOURCE_MANAGER),$(PROTO_DIR)/grpc_resource_alloc_procotol/*.proto,remote-runner)

build-proto-python:
	cd $(PYTHON_PROJECT_ROOT) && $(PYTHON) -m grpc_tools.protoc -I=$(PROTO_DIR)/grpc_ce_protocol --python_out=$(PYTHON_PROTO) --grpc_python_out=$(PYTHON_PROTO) $(PROTO_DIR)/grpc_ce_protocol/*.proto && \
	sed -i 's/import ceprotocol_pb2/from sdk.grpc_ce_protocol import ceprotocol_pb2/g' $(PYTHON_PROTO)/*.py
build-proto:
	make build-proto-go
	make build-proto-python

clean-sdk-python:
	rm -rf $(PYTHON_PROJECT_ROOT)/dist
	rm -rf $(PYTHON_PROJECT_ROOT)/build
	rm -rf $(PYTHON_PROJECT_ROOT)/sdk.egg-info

clean-node-testing:
	rm -rf $(NODE_TESTING_DIR)

move-sdk-dist:
	cp $(DISTRIBUTION_DIR)/*.whl $(GO_PROJECT_REMOTE_RUNNER)/py_src

build-sdk:
	make clean-dist
	make clean-node-testing
	mkdir -p $(DISTRIBUTION_DIR)
	cd $(PYTHON_PROJECT_ROOT) && $(PYTHON) setup.py sdist bdist_wheel
	cp $(PYTHON_PROJECT_ROOT)/dist/* $(DISTRIBUTION_DIR)
	make clean-sdk-python
	make move-sdk-dist

build: 
	make build-proto
	make build-sdk

clean-dist:
	rm -rf $(DISTRIBUTION_DIR)

run-resource-manager:
	cd $(GO_PROJECT_RESOURCE_MANAGER) && $(GO) run .

run-remote-runner-local:
	cd $(GO_PROJECT_REMOTE_RUNNER) && $(GO) run . local-1 small-4cpu-8gb localhost:50050

run-server:
	cd $(GO_PROJECT_CLIENT_ENGING) && $(GO) run .

run-client-example:
	cd $(PYTHON_PROJECT_ROOT) && $(PYTHON) main.py
