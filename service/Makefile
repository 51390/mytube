.phony: clean all format sync clean

all: bin/sync bin/server

SOURCES := $(shell find . -iname '*.go')

bin/sync: bin $(SOURCES)
	go build -o bin/sync cmd/sync/main.go

bin/server: bin $(SOURCES)
	go build -o bin/server cmd/server/main.go

bin:
	mkdir bin

format: cmd yourtube
	gofmt -w cmd yourtube

clean:
	rm -rf bin

sync:
	./bin/sync
