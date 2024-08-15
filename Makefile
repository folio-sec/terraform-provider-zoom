TEST?=$$(go list ./... |grep -v 'vendor')
DEV      := folio-sec
PROVIDER := zoom
# VERSION  := $(shell git describe --abbrev=0 --tags --match "v*")
VERSION := v0.0.1
PLUGINS  := ${HOME}/bin/plugins/registry.terraform.io/${DEV}/${PROVIDER}
BIN      := terraform-provider-zoom_${VERSION}

define TERRAFORMRC

add the following config to ~/.terraformrc to enable override:
```
provider_installation {
  dev_overrides {
    "${DEV}/${PROVIDER}" = "${PLUGINS}"
  }
}
```
endef

.PHONY: tools
tools:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin

.PHONY: tidy
tidy:
	@go mod tidy -v

.PHONY: install
install:
	@go mod download

.PHONY: format
format:
	@golangci-lint run --fix ./...

.PHONY: lint
lint:
	@golangci-lint run -v ./...

.PHONY: test
test:
	@go vet ./...
	go test $(TEST) -race -v -shuffle on

test/%:
	@go vet ./$(@:test/%=%)
	go test -race -v -shuffle on ./$(@:test/%=%)

.PHONY: testacc
testacc:
	TF_ACC=1 go test $(TEST) -race -v $(TESTARGS) -shuffle on -ldflags="-X=github.com/folio-sec/terraform-provider-zoom/version.ProviderVersion=acc"

.PHONY: generate
generate:
	@go generate ./...

# Run go build. Output to dist/.
.PHONY: build
build:
	@mkdir -p dist
	go build -o dist/${BIN} .

# Run go build. Output to dist/.
.PHONY: build_override
build_override: build
	mkdir -p ${PLUGINS}
	mv dist/${BIN} ${PLUGINS}/${BIN}

# Run go build. Move artifact to terraform plugins dir. Output override config for ~/.terraformrc
.PHONY: local_install
local_install: build_override
	$(info ${TERRAFORMRC})

.PHONY: updatespec
updatespec: updatespec/phone

.PHONY: updatespec/phone
updatespec/phone:
	@curl -sfL https://developers.zoom.us/api-specs/phone/methods/ZoomPhoneAPI-spec.json | ./scripts/patchSpec.js > spec/ZoomPhoneAPISpec.json

.PHONY: updatespec/user
updatespec/user:
	@curl -sfL https://developers.zoom.us/api-specs/user/methods/ZoomUserAPI-spec.json | ./scripts/patchSpec.js > spec/ZoomUserAPISpec.json
