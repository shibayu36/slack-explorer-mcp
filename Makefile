.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	go test -v ./...
