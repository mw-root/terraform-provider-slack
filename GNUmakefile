default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./... ;\
    cp ${HOME}/go/bin/terraform-provider-slack ${HOME}/.terraform.d/plugins/terraform.local.com/mw-root/slack/0.0.1/darwin_arm64/terraform-provider-slack

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: fmt lint test testacc build install generate
