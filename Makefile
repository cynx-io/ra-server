tidy:
	go mod tidy
	go fmt ./internal/...
	fieldalignment -fix ./internal/...
	go vet ./internal/...
	golangci-lint run --fix ./internal/...
	staticcheck ./internal/...

	go fmt ./main.go
	fieldalignment -fix ./main.go
	go vet ./main.go
	golangci-lint run --fix ./main.go
	staticcheck ./main.go

run:
	make clean
	make proto
	make tidy
	go run main.go

install_deps:
	# These needs sudo
	# apt install build-essential -y
    # curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/google/wire/cmd/wire@latest
	go get -u gorm.io/gorm
	go get -u gorm.io/driver/sqlite


#proto-gen:
#	@echo "Generating proto files..."
#	@for service in $(MICROSERVICES); do \
#		echo "Processing $$service..."; \
#		mkdir -p $(PROTO_GEN_DIR)/$$service; \
#		protoc \
#			-I=$(PROTO_SRC_DIR) \
#			-I=$(PROTO_SRC_DIR)/../third_party \
#			--go_out=$(PROTO_GEN_DIR)/$$service \
#			--go_opt=paths=source_relative \
#			--go-grpc_out=$(PROTO_GEN_DIR)/$$service \
#			--go-grpc_opt=paths=source_relative \
#			--validate_out=lang=go:$(PROTO_GEN_DIR)/$$service \
#			--validate_opt=paths=source_relative \
#			--experimental_allow_proto3_optional \
#			$(PROTO_SRC_DIR)/$$service/*.proto; \
#	done

# Configuration
PROTO_SRC_DIR := api/proto/src
PROTO_GEN_DIR := api/proto/gen
MICROSERVICES := $(notdir $(wildcard $(PROTO_SRC_DIR)/*))

# Proto generation
proto-clean:
	@echo "Cleaning generated proto files..."
	rm -rf $(PROTO_GEN_DIR)/*

proto-gen:
	@echo "Generating proto files..."
	cd . && buf generate

proto: proto-clean proto-gen

.PHONY: clean
clean:
	@echo "Cleaning generated files..."
	rm -f api/proto/gen/ra/ra/*.pb.go
	rm -f bin/ra

.PHONY: all
all: clean proto build

.phony: update-core
update-core:
	@echo "Updating core dependencies..."
	go get -u github.com/cynxees/cynx-core@v0.0.26

build_and_push:
	docker buildx build \
		--platform linux/amd64 \
		--tag derwin334/ra-server-dev:latest \
		--push \
		.
