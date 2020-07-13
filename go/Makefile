PACKAGE := github.com/opencredo/venafi-cloud-ab-poc/go
SWAGGER_VERSION := "3.28.0"

.DEFAULT_GOAL := all
.PHONY := clean all fmt linux coverage release build swagger-ui

release_dir := bin/release/
go_files := $(shell find . -path ./vendor -prune -o -path '*/testdata' -prune -o -name '*.gen.go' -prune -o -path '*/statik' -prune -o -type f -name '*.go' -print)
commands := $(notdir $(shell find cmd/* -type d -prune))
local_bins := $(addprefix bin/,$(commands))
GIT_COMMIT_ID ?= $(shell git rev-list -1 HEAD)
API_DIR := $(shell if [ -e .api ]; then echo .api; else echo ../api; fi)

clean:
	# Removing all generated files...
	@rm -rf bin/ || true
	@find ./internal/ -name '*.gen.go' -exec rm {} \; || true

bin/.vendor: go.mod go.sum
	# Downloading modules...
	@go mod download
	@mkdir -p bin/
	@touch bin/.vendor

bin/.generate: $(go_files) bin/.vendor
	# Generating files
	@go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen
	@oapi-codegen -generate types,chi-server,spec -package ledgerserver ${API_DIR}/ledger.yaml > internal/pkg/ledgerserver/api.gen.go
	@oapi-codegen -generate client -package ledgerserver ${API_DIR}/ledger.yaml > internal/pkg/ledgerserver/client.gen.go
	@go generate ./...
	@touch bin/.generate

fmt: bin/.generate $(go_files)
	# Formatting files...
	@go run golang.org/x/tools/cmd/goimports -w $(go_files)

bin/.vet: bin/.generate $(go_files)
	go vet  ./...
	@touch bin/.vet

bin/.fmtcheck: bin/.generate $(go_files)
	# Checking format of Go files...
	@GOIMPORTS=$$(go run golang.org/x/tools/cmd/goimports -l $(go_files)) && \
	if [ "$$GOIMPORTS" != "" ]; then \
		go run golang.org/x/tools/cmd/goimports -d $(go_files); \
		exit 1; \
	fi
	@touch bin/.fmtcheck

bin/.coverage.out: bin/.generate $(go_files)
	@go test -cover -v -count=1 ./... -coverpkg=$(shell go list ${PACKAGE}/... | xargs | sed -e 's/ /,/g') -coverprofile bin/.coverage.tmp
	@mv bin/.coverage.tmp bin/.coverage.out

coverage: bin/.coverage.out
	@go tool cover -html=bin/.coverage.out

$(local_bins): bin/.fmtcheck bin/.vet bin/.coverage.out $(go_files)
	CGO_ENABLED=0 go build -ldflags "-X main.GitCommit=$(GIT_COMMIT_ID)" -o $@ $(PACKAGE)/cmd/$(basename $(@F))

build: $(local_bins)

all: build

swagger-ui:
	@go get github.com/rakyll/statik
	@curl -L -o /tmp/swagger.zip https://github.com/swagger-api/swagger-ui/archive/v${SWAGGER_VERSION}.zip
	@cd /tmp && unzip swagger.zip
	@cd /tmp/swagger-ui-${SWAGGER_VERSION}/dist && sed -i 's/https:\/\/petstore.swagger.io\/v2\/swagger.json/.\/swaggerui\/swagger.json/g' index.html
	@mkdir -p internal/pkg/swaggerui
	@rm -rf internal/pkg/swaggerui/statik
	@cd internal/pkg/swaggerui && statik -src /tmp/swagger-ui-${SWAGGER_VERSION}/dist 
	@rm /tmp/swagger.zip && rm -r /tmp/swagger-ui-${SWAGGER_VERSION}
