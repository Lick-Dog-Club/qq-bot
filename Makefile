LDFLAGS=-w -s

.PHONY: build-push
build-push:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o app-linux-amd64 . && \
	docker build --platform linux/amd64  -t duccnzj/qq-bot . && docker push duccnzj/qq-bot && \
	rm app-linux-amd64

.PHONY: build-linux-amd64
build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o app-linux-amd64 .
