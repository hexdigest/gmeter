all: test lint

lint:
	golint ./... && go vet ./...

test:
	go test -race ./...

install:
	go install ./cmd/gmeter

release:
	GOOS=windows GOARCH=386 go build -o bin/gmeter-windows-i386-v$(version).exe cmd/gmeter/gmeter.go
	GOOS=windows GOARCH=amd64 go build -o bin/gmeter-windows-amd64-v$(version).exe cmd/gmeter/gmeter.go
	GOOS=linux GOARCH=386 go build -o bin/gmeter-linux-i386-v$(version) cmd/gmeter/gmeter.go
	GOOS=linux GOARCH=amd64 go build -o bin/gmeter-linux-amd64-v$(version) cmd/gmeter/gmeter.go
	GOOS=darwin GOARCH=amd64 go build -o bin/gmeter-darwin-amd64-v$(version) cmd/gmeter/gmeter.go

	zip -m bin/gmeter-darwin-amd64-v$(version).zip bin/gmeter-darwin-amd64-v$(version)
	zip -m bin/gmeter-windows-amd64-v$(version).zip bin/gmeter-windows-amd64-v$(version).exe
	zip -m bin/gmeter-windows-i386-v$(version).zip bin/gmeter-windows-i386-v$(version).exe
	gzip bin/gmeter-linux-i386-v$(version)
	gzip bin/gmeter-linux-amd64-v$(version)

all: lint test install

