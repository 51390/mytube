.phony: .env run-service run-app init-app build-service

all: .env

run-app: .env
	flask -e .env --app app run --debug

init-app:
	echo 'db.create_all()' | flask --app app shell

run-service: build-service .env
	./service/bin/server

build-service:
	make -C service bin/server

.env:
	@echo "POSTGRES_PASSWORD=$(shell head -n 1024 /dev/urandom | md5sum | sed 's/ .*//g')" > .env

