.phony: run init-db

run:
	flask  --app . run --debug

init-db:
	echo 'db.create_all()' | flask --app . shell
