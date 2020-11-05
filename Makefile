VERSION ?= local
GIT_SHA ?= $(shell git rev-parse HEAD)
TIMESTAMP ?= $(shell date +%Y-%m-%dT%T)
LD_FLAGS := -X main.version=${VERSION} -X main.commit=${GIT_SHA} -X main.date=${TIMESTAMP}

build-deps:
	GO111MODULE=off go get -u golang.org/x/lint/golint
	GO111MODULE=off go get -u oss.indeed.com/go/go-groups

deps:
	go mod download
	go mod verify

fmt:
	goimports -w .
	go-groups -w .
	gofmt -s -w .

test:
	go vet ./...
	golint -set_exit_status ./...
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

build:
	go build -ldflags="$(LD_FLAGS)" ./cmd/vcs-connect/

docker:
	docker build . \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_SHA=$(GIT_SHA) \
		--build-arg TIMESTAMP=$(TIMESTAMP) \
		--tag effxhq/vcs-connect:latest \
		--tag effxhq/vcs-connect:$(VERSION) \
		-f Dockerfile

dockerx:
	docker buildx build . \
		--platform linux/amd64,linux/arm64,linux/arm/v7 \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_SHA=$(GIT_SHA) \
		--build-arg TIMESTAMP=$(TIMESTAMP) \
		--tag effxhq/vcs-connect:latest \
		--tag effxhq/vcs-connect:$(VERSION) \
		-f Dockerfile \
		--push
