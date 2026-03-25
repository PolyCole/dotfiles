.PHONY: all startup-message clean

all: startup-message

startup-message:
	go build -o bin/startup-message tools/startup-message/main.go

clean:
	rm -f bin/startup-message
