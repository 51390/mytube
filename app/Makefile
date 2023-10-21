phony: init-app run-app

run-app-dev:
	flask --app . run --debug --host=0.0.0.0

init-app:
	pip install -r requirements.txt

init-db:
	PGPASSWORD=$$POSTGRES_PASSWORD psql -h $$POSTGRES_HOST -U $$POSTGRES_USER -c 'create database mytube_app;'
	echo 'db.create_all()' | flask --app . shell
