.PHONY: all startup-message dots test clean

all: startup-message dots

startup-message:
	go build -o bin/startup-message tools/startup-message/main.go

dots:
	cd tools/dots && go build -o ../../bin/dots ./cmd/dots

test:
	cd tools/dots && go vet ./... && go test ./...

clean:
	rm -f bin/startup-message bin/dots
