.PHONY: start
start:
	@echo "starting containers"
	sudo -E docker-compose up
.PHONY: start-detached
start-detached:
	@echo "starting containers in background"
	sudo -E docker-compose up -d
.PHONY: stop
stop:
	@echo "stopping containers"
	docker-compose stop
.PHONY: down
down:
	@echo "stopping and removing containers and networks"
	docker-compose down
.PHONY: clean
clean:
	@echo "stopping and removing containers, associated volumes and networks"
	docker-compose down -v
.PHONY: restart
restart:
	@echo "restarting containers"
	docker-compose restart
.PHONY: logs
logs:
	./logs $t