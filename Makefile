.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	go test -v ./... -timeout 5m
