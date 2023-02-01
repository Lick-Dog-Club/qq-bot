VERSION_PATH=$(shell go list -m -f "{{.Path}}")/sys-update
LDFLAGS=-w -s \
-X ${VERSION_PATH}.gitCommit=$(shell git rev-parse --short HEAD)

.PHONY: build-push
build-push:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o app-linux-amd64 . && \
	docker build --platform linux/amd64  -t duccnzj/qq-bot . && docker push duccnzj/qq-bot && \
	rm app-linux-amd64

.PHONY: build-linux-amd64
build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o app-linux-amd64 .

.PHONY: build
build:
	go build -ldflags="${LDFLAGS}" -o app .
