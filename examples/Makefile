PKGS        := `go list ./...`
LDFLAGS 	:=-ldflags "-s -w "

.PHONY: clean

run-osx: build-osx
	./gofre_examples_osx

run: build
	./gofre_examples_linux

build-osx: clean
	CGO_ENABLED=0 GO111MODULE=on go build ${LDFLAGS} -a -o gofre_examples_osx ./*.go

build: clean
	CGO_ENABLED=0 GO111MODULE=on GOOS=linux go build ${LDFLAGS} -a -o gofre_examples_linux ./*.go

clean:
	go clean