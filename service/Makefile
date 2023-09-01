.phony: clean

bin/sync: bin cmd/sync/main.go
	go build -o bin/sync cmd/sync/main.go

bin:
	mkdir bin

clean:
	rm -rf bin
