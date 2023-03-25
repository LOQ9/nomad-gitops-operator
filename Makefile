.PHONY: test build compile start-nomad

test:
	go test -v ./...

build:
	go build -o nomad-gitops-operator

compile:
	echo "Compiling for every OS and Platform"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/nomad-gitops-operator-linux-amd64 main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o bin/nomad-gitops-operator-linux-arm main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/nomad-gitops-operator-linux-arm64 main.go
	CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -o bin/nomad-gitops-operator-freebsd-386 main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o bin/nomad-gitops-operator.exe main.go

start-nomad:
	./scripts/start-nomad.sh

all: test build