PKGS        := `go list ./...`
LDFLAGS 	:=-ldflags "-s -w "

.PHONY: clean fmt vet staticcheck test

build-osx: clean fmt vet staticcheck test
	CGO_ENABLED=0 GO111MODULE=on go build ${LDFLAGS} -a ./*.go

build: clean fmt vet staticcheck test
	CGO_ENABLED=0 GO111MODULE=on GOOS=linux go build ${LDFLAGS} -a ./*.go

clean:
	go clean

staticcheck:
	staticcheck -tests=false ./...

fmt:
	find . -type f -name '*.go' | xargs gofmt -w -s

vet:
	go vet $(PKGS)

test:
	go test ./...