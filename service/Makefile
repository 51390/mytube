.phony: clean all format sync clean run

all: bin/sync bin/server

SOURCES := $(shell find . -iname '*.go')

bin/sync: bin $(SOURCES)
	go build -o bin/sync cmd/sync/main.go

bin/server: bin $(SOURCES)
	echo $$GOPATH $(shell pwd)
	go build -o bin/server cmd/server/main.go

bin:
	mkdir bin

format: cmd mytube
	gofmt -w cmd mytube

clean:
	rm -rf bin

sync:
	./bin/sync

init-db:
	-PGPASSWORD=$$POSTGRES_PASSWORD psql -h $$POSTGRES_HOST -U $$POSTGRES_USER -c 'create database mytube_service;'
	PGPASSWORD=$$POSTGRES_PASSWORD psql -h $$POSTGRES_HOST -U $$POSTGRES_USER --dbname=mytube_service -c 'select current_database();'

run:
	./bin/server
