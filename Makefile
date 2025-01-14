.PHONY: all clean

# Variables
RELEASE_VERSION := 0.0.1

# --- Files and directories --- #
# Harpy components
GO_PROJECT_CLIENT_ENGING := $(shell pwd)/client-engine
GO_PROJECT_REMOTE_RUNNER := $(shell pwd)/remote-runner
GO_PROJECT_RESOURCE_MANAGER := $(shell pwd)/resource-manager
# Python SDK
PYTHON_PROJECT_DIR := $(shell pwd)/pyharpy/harpy
PYTHON_PROJECT_ROOT := $(shell pwd)/pyharpy
# Docker compose directory (this will set up the environment for testing)
COMPOSE_ROOT := $(shell pwd)/compose_engine
# Protobuf files
PROTO_DIR := $(shell pwd)/grpc_protos
PYTHON_PROTO := $(PYTHON_PROJECT_DIR)/grpc_ce_protocol
GRPC_PYTHON_PLUGIN=$(shell $(PYTHON) -m grpc_tools.protoc --plugin)
NODE_LOCAL_TESTING_DIR := $(GO_PROJECT_REMOTE_RUNNER)/py_src/venv
NODE_LOCAL_TESTING_BLOCK_DIR := $(GO_PROJECT_REMOTE_RUNNER)/_block

# --- Commands --- #
# Compiler 
GO := GODEBUG=gctrace=1 go
PYTHON := poetry run python
GOFLAGS := -buildmode=c-shared
# Docker commands
DOCKER := sudo docker
DOCKER_COMPOSE := sudo docker compose
DISTRIBUTION_DIR := $(shell pwd)/dist


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
	rm -r $(PYTHON_PROTO)
	mkdir $(PYTHON_PROTO)
	touch $(PYTHON_PROTO)/__init__.py
	cd $(PYTHON_PROJECT_ROOT) && $(PYTHON) -m grpc_tools.protoc -I=$(PROTO_DIR)/grpc_ce_protocol --python_out=$(PYTHON_PROTO) --grpc_python_out=$(PYTHON_PROTO) $(PROTO_DIR)/grpc_ce_protocol/*.proto && \
	sed -i 's/import ceprotocol_pb2/from harpy.grpc_ce_protocol import ceprotocol_pb2/g' $(PYTHON_PROTO)/*.py
build-proto:
	make build-proto-go
	make build-proto-python

# Build the SDK binaries
clean-sdk-python:
	rm -rf $(PYTHON_PROJECT_ROOT)/dist
	rm -rf $(PYTHON_PROJECT_ROOT)/build
	rm -rf $(PYTHON_PROJECT_ROOT)/sdk.egg-info
clean-node-testing:
	rm -rf $(NODE_LOCAL_TESTING_DIR)
	rm -rf $(NODE_LOCAL_TESTING_BLOCK_DIR)
move-sdk-dist:
	cp $(DISTRIBUTION_DIR)/*.whl $(GO_PROJECT_REMOTE_RUNNER)/py_src
clean-dist:
	rm -rf $(DISTRIBUTION_DIR)
build-sdk:
	make clean-dist
	make clean-node-testing
	mkdir -p $(DISTRIBUTION_DIR)
	cd $(PYTHON_PROJECT_ROOT) && $(PYTHON) setup.py sdist bdist_wheel
	cp $(PYTHON_PROJECT_ROOT)/dist/* $(DISTRIBUTION_DIR)
	make clean-sdk-python
	make move-sdk-dist

# Build the docker images
build-docker-images:
	make clean-node-testing
	$(DOCKER) system prune -f
	$(DOCKER) system prune --volumes -f
	$(DOCKER) build -t harpy-base-go-python:$(RELEASE_VERSION) -f ./images/base-image/Dockerfile ./images/base-image/
	$(DOCKER) build -t harpy:$(RELEASE_VERSION) -f ./images/harpy-image/Dockerfile .
	$(DOCKER) build -t harpy-jupyter-server:$(RELEASE_VERSION) -f ./images/harpy-jupyter-server/Dockerfile .

# Prepare the environment
prepare-env:
	cd $(PYTHON_PROJECT_ROOT) && poetry install
# Docker compose instructions
run-compose:
	cd $(COMPOSE_ROOT) && $(DOCKER_COMPOSE) up
run-compose-down:
	cd $(COMPOSE_ROOT) && $(DOCKER_COMPOSE) down

build:
	make run-compose-down
	make prepare-env
	make build-proto
	make build-sdk
	make build-docker-images

run-export-images:
	$(DOCKER) save harpy:$(RELEASE_VERSION) > harpy.tar
	$(DOCKER) save harpy-jupyter-server:$(RELEASE_VERSION) > harpy-jupyter-server.tar

# These are testing calls (not used in the final version) - these may be unstable depending on how the environment is set up 
# Prefer to use the docker images instead
run-dev-resource-manager:
	cd $(GO_PROJECT_RESOURCE_MANAGER) && $(GO) run .
run-dev-remote-runner-local:
	cd $(GO_PROJECT_REMOTE_RUNNER) && $(GO) run . local-1 localhost:50050
run-dev-server:
	cd $(GO_PROJECT_CLIENT_ENGING) && $(GO) run .
run-dev-client-example:
	cd $(PYTHON_PROJECT_ROOT) && $(PYTHON) main.py

ifneq ($(K_FLAG),)
    K_FLAG_P := -k $(K_FLAG)
endif
run-integrety-test:
	cd $(PYTHON_PROJECT_ROOT) && $(PYTHON) -m pytest $(K_FLAG_P) -v ../test/. 

dev-nuke-all:
	sudo fuser -k 50050/tcp
	sudo fuser -k 50051/tcp
	sudo fuser -k 50052/tcp
	sudo fuser -k 50053/tcp
	