TEST?=$$(go list ./... |grep -v 'vendor')

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
