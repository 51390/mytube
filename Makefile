.phony: .env run-service run-app init-app build-service run-db db-down

MD5 := $(shell test `uname` = Linux && echo md5sum || echo md5)

all: .env

build: .env
	docker compose build

build-minikube: .env
	eval `minikube docker-env`; make build

run: .env
	docker compose up db -d
	sleep 3
	docker compose up

down: .env
	docker compose down

kubectl-config-aws:
	aws eks update-kubeconfig --region sa-east-1 --name mytube

kubectl-config-minukube:
	minikube start

kubectl-stop-minikube:
	minikube stop

kubectl-namespace:
	kubectl create namespace mytube

kubectl-secrets: .env.kubernetes
	kubectl -n mytube create secret generic credentials --from-env-file=.env.kubernetes

kubectl-deployments:
	kubectl apply -f kubernetes/app-deployment.yml
	kubectl apply -f kubernetes/service-deployment.yml

kubectl-db-deployment:
	kubectl apply -f kubernetes/db-deployment.yml

kubectl-connetivity:
	kubectl create -f kubernetes/service-cluster-ip.yml -n mytube
	kubectl create -f kubernetes/app-load-balancer.yml -n mytube
	kubectl create -f kubernetes/db-cluster-ip.yml -n mytube

.env.kubernetes: .env
	cat .env | sed 's/="/=/g' | sed 's/"$$//g' > $@

.env:
	@echo "POSTGRES_PASSWORD=\"$(shell head -n 1024 /dev/urandom | $(MD5) | sed 's/ .*//g')\"" > .env
	@echo 'POSTGRES_USER="postgres"' >> .env
	@echo 'POSTGRES_HOST="db"' >> .env
	@echo 'POSTGRES_SSLMODE="allow"' >> .env
	@echo "JWT_TOKEN_SECRET=\"$(shell head -n 1024 /dev/urandom | $(MD5) | sed 's/ .*//g')\"" >> .env
	@cat  client_secret.json | jq -c '.["installed"]' | tr ',' '\n' | sed 's/[{}]//g' | sed 's/^"//g' | sed 's/":/ /g' | awk '{ print toupper($$1) "=" $$2}' >> .env
