NAME=dbtcloud
BINARY=terraform-provider-$(NAME)

default: install

setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	go install golang.org/x/tools/cmd/goimports@latest

build:
	go build -ldflags "-w -s" -o $(BINARY) .

install: build
	mkdir -p $(HOME)/.terraform.d/plugins
	mv ./$(BINARY) $(HOME)/.terraform.d/plugins/$(BINARY)

doc:
	go generate ./... && rm docs/resources/dbt_cloud_* && rm docs/data-sources/dbt_cloud_* && cp -r guides docs/

test: deps
	go test -mod=readonly -count=1 ./...

test-acceptance: deps
	TF_ACC=1 TEST_DATABRICKS=false go test -v -mod=readonly -count=1 ./...

check-docs: docs
	git diff --exit-code -- docs

deps:
	go mod tidy

release:
	git tag "v$(VERSION)"
	git push origin "v$(VERSION)"
