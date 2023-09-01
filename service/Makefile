.phony: clean

SOURCES := $(shell find . -iname '*.go')

bin/sync: bin $(SOURCES)
	go build -o bin/sync cmd/sync/main.go

bin:
	mkdir bin

clean:
	rm -rf bin
