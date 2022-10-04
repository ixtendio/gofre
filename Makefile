PKGS        := `go list ./...`
LDFLAGS 	:=-ldflags "-s -w "

.PHONY: clean fmt vet test package

run-osx: clean fmt vet test build-osx
	./gow_examples

run: clean fmt vet test build
	./gow_examples

build-osx: clean fmt vet test
	CGO_ENABLED=0 GO111MODULE=on go build ${LDFLAGS} -a -o gow_examples examples/*.go

build: clean fmt vet test
	CGO_ENABLED=0 GO111MODULE=on GOOS=linux go build ${LDFLAGS} -a -o gow_examples examples/*.go

clean:
	go clean

fmt:
	find . -type f -name '*.go' | xargs gofmt -w -s

vet:
	go vet $(PKGS)

test:
	go test ./...