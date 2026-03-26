.PHONY: all startup-message dots clean

all: startup-message dots

startup-message:
	go build -o bin/startup-message tools/startup-message/main.go

dots:
	cd tools/dots && go build -o ../../bin/dots ./cmd/dots

clean:
	rm -f bin/startup-message bin/dots
