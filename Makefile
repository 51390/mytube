.phony: .env run-service run-app init-app build-service

all: .env

run-app: .env
	flask -e .env --app app run --debug

init-app: .env
	echo 'db.create_all()' | flask --app app shell

run-service: build-service .env
	./service/bin/server

build-service:
	make -C service bin/server

.env:
	@echo "POSTGRES_PASSWORD=$(shell head -n 1024 /dev/urandom | md5sum | sed 's/ .*//g')" > .env
	@cat  client_secret.json | jq -c '.["installed"]' | tr ',' '\n' | sed 's/[{}]//g' | sed 's/^"//g' | sed 's/":/ /g' | awk '{ print toupper($$1) "=" $$2}' >> .env

