.phony: .env run-service run-app init-app build-service run-db db-down

MD5 := $(shell test `uname` = Linux && echo md5sum || echo md5)

all: .env

run-db: .env
	docker compose up

db-down: .env
	docker compose down

run-app: .env
	flask -e .env --app app run --debug

init-app: .env
	pip install -r app/requirements.txt
	echo 'db.create_all()' | flask -e .env --app app shell

run-service: build-service .env
	./service/bin/server

build-service:
	make -C service bin/server

.env:
	@echo "POSTGRES_PASSWORD=$(shell head -n 1024 /dev/urandom | $(MD5) | sed 's/ .*//g')" > .env
	@cat  client_secret.json | jq -c '.["installed"]' | tr ',' '\n' | sed 's/[{}]//g' | sed 's/^"//g' | sed 's/":/ /g' | awk '{ print toupper($$1) "=" $$2}' >> .env

