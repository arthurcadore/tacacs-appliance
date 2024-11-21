all: stop start

stop:
	docker compose down

start: build
	docker compose up &
	
build:
	docker compose build --no-cache

clean: stop
	docker ps -a -q | xargs docker rm ; docker images -q | xargs docker rmi -f ; docker volume ls -q | xargs docker volume rm