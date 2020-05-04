REPO=github-rulla-nycklar
CONTAINER=quay.io/shelmangroup/github-rulla-nycklar
VERSION ?= $(shell ./hacks/git-version)
LD_FLAGS="-X main.Version=$(VERSION) -w -s -extldflags \"-static\" "

$( shell mkdir -p _bin )
$( shell mkdir -p _release )

default: format format-verify build-dev

clean:
	@rm -r _bin _release

test: format-verify
	@echo "----- running tests -----"
	@go test -v -i $(shell go list ./... | grep -v '/vendor/')
	@go test -v $(shell go list ./... | grep -v '/vendor/')

install:
	@GOBIN=$(GOPATH)/bin && go install -mod=readonly -v -ldflags $(LD_FLAGS) 

build-dev:
	@echo "----- running dev build-----"
	@export GOBIN=$(PWD)/_bin && go install -v -mod=readonly -ldflags $(LD_FLAGS) 

build:
	@echo "----- running release build -----"
	@go build -v -o _release/$(REPO) -ldflags $(LD_FLAGS) 

container:
	@docker build -t $(CONTAINER):$(VERSION) --file Dockerfile .

container-push:
	@docker push $(CONTAINER):$(VERSION)

deploy:
	@sed -e "s/{{VERSION}}/${VERSION}/g;" template/deployment.yaml > deployment.yaml
	@kubectl apply -f deployment.yaml

download:
	@go mod download

setup:
	@go get -u golang.org/x/tools/cmd/goimports

format:
	@echo "----- running gofmt -----"
	@gofmt -w -s *.go
	@echo "----- running goimports -----"
	@goimports -w *.go

format-verify:
	@echo "----- running gofmt verify -----"
	@hacks/verify-gofmt
	@echo "----- running goimports verify -----"
	@hacks/verify-goimports

