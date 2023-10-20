phony: init-app run-app

run-app-dev:
	flask --app . run --debug --host=0.0.0.0

init-app:
	pip install -r requirements.txt

init-db:
	echo 'db.create_all()' | flask --app . shell
