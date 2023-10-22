MD5 := $(shell test `uname` = Linux && echo md5sum || echo md5)
ENVVARS := source .env; source .versions
TERRAFORM_VARS := -var="POSTGRES_PASSWORD=$$POSTGRES_PASSWORD" -var="APP_VERSION=$$APP_VERSION" -var="SERVICE_VERSION=$$SERVICE_VERSION" -var="DB_VERSION=$$DB_VERSION"

.env:
	@echo "POSTGRES_PASSWORD=\"$(shell head -n 1024 /dev/urandom | $(MD5) | sed 's/ .*//g')\"" > .env
	@echo 'POSTGRES_USER="postgres"' >> .env
	@echo 'POSTGRES_HOST="db"' >> .env
	@echo 'POSTGRES_SSLMODE="allow"' >> .env
	@echo "JWT_TOKEN_SECRET=\"$(shell head -n 1024 /dev/urandom | $(MD5) | sed 's/ .*//g')\"" >> .env
	@cat  client_secret.json | jq -c '.["installed"]' | tr ',' '\n' | sed 's/[{}]//g' | sed 's/^"//g' | sed 's/":/ /g' | awk '{ print toupper($$1) "=" $$2}' >> .env

.env.kubernetes: .env
	cat .env | sed 's/="/=/g' | sed 's/"$$//g' > $@

.env.aws: .env
	cat .env | sed 's/="/=/g' | sed 's/"$$//g' > $@

docker-images:
	eval `minikube docker-env`; docker pull postgres
	eval `minikube docker-env`; docker pull ubuntu

docker-login:
	eval `minikube docker-env`; docker login

compose-build: .env
	docker compose --env-file .versions build 

compose-run: .env
	docker compose up db -d
	sleep 3
	docker compose up db-init -d
	sleep 3
	docker compose up

compose-down: .env
	docker compose down

helm-versions: infrastructure/helm/mytube/versions.yml

helm-terraform-values:
	make terraform-output | sed 's/ = /: /g' > infrastructure/helm/mytube/terraform-values.yml

helm-install: minikube-build helm-versions
	helm install mytube-release ./infrastructure/helm/mytube --namespace mytube \
		-f infrastructure/helm/mytube/versions.yml \
		-f infrastructure/helm/mytube/values.yml

helm-install-aws: kubectl-config-aws kubectl-namespace kubectl-secrets helm-versions helm-terraform-values
	helm install mytube-release ./infrastructure/helm/mytube --namespace mytube \
		-f infrastructure/helm/mytube/versions.yml \
		-f infrastructure/helm/mytube/values.yml \
		-f infrastructure/helm/mytube/terraform-values.yml

helm-upgrade: minikube-build helm-versions
	helm upgrade mytube-release ./infrastructure/helm/mytube --namespace mytube \
		-f infrastructure/helm/mytube/versions.yml \
		-f infrastructure/helm/mytube/values.yml

helm-upgrade-aws: kubectl-config-aws helm-versions helm-terraform-values
	helm upgrade mytube-release ./infrastructure/helm/mytube --namespace mytube \
		-f infrastructure/helm/mytube/versions.yml \
		-f infrastructure/helm/mytube/values.yml \
		-f infrastructure/helm/mytube/terraform-values.yml

helm-uninstall:
	helm uninstall mytube-release --namespace mytube

helm-uninstall-aws: kubectl-config-aws
	helm uninstall mytube-release --namespace mytube

infrastructure/helm/mytube/versions.yml: .versions
	cat .versions | sed 's/=/: /g' | awk '{ print tolower($0) }' > $@

kubectl-config-aws:
	aws eks update-kubeconfig --region sa-east-1 --name mytube

kubectl-config-minikube:
	minikube start

kubectl-namespace:
	-kubectl create namespace mytube

kubectl-secrets: .env.kubernetes
	-kubectl -n mytube delete secret credentials
	kubectl -n mytube create secret generic credentials --from-env-file=.env.kubernetes

kubectl-secrets-minikube: .env.minikube
	-kubectl -n mytube delete secret credentials
	kubectl -n mytube create secret generic credentials --from-env-file=.env.minikube

kubectl-deployments:
	kubectl apply -f kubernetes/app-deployment.yml
	kubectl apply -f kubernetes/service-deployment.yml

kubectl-db-deployment:
	kubectl apply -f kubernetes/db-deployment.yml
	kubectl apply -f kubernetes/db-init-job.yml

kubectl-connectivity:
	kubectl create -f kubernetes/service-cluster-ip.yml -n mytube
	kubectl create -f kubernetes/app-load-balancer.yml -n mytube
	kubectl create -f kubernetes/db-cluster-ip.yml -n mytube

minikube-build: .env
	eval `minikube docker-env`; make compose-build

minikube-images:
	eval `minikube docker-env`; docker images

minikube-tunnel:
	minikube tunnel -c

minikube-stop:
	minikube stop

minikube-bootstrap: kubectl-config-minikube kubectl-namespace kubectl-secrets-minikube helm-install

minikube-load-balancer:
	minikube service mytube-app-load-balancer -n mytube

minikube-teardown:
	-kubectl delete namespace mytube
	make minikube-stop

terraform-init:
	terraform -chdir=./infrastructure/terraform/aws init

terraform-plan: .env
	$(ENVVARS) ; terraform -chdir=./infrastructure/terraform/aws plan $(TERRAFORM_VARS)

terraform-apply: .env compose-build
	$(ENVVARS) ; terraform -chdir=./infrastructure/terraform/aws apply $(TERRAFORM_VARS)

terraform-destroy: .env
	$(ENVVARS) ; terraform -chdir=./infrastructure/terraform/aws destroy $(TERRAFORM_VARS)

terraform-output:
	@terraform -chdir=./infrastructure/terraform/aws output

