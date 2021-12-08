.PHONY: start
start:
	@echo "starting dp-compose containers"
	docker-compose up

.PHONY: start-detached
start-detached:
	@echo "starting dp-compose containers in background"
	docker-compose up -d

.PHONY: stop
stop:
	@echo "stopping dp-compose containers"
	docker-compose stop

.PHONY: down
down:
	@echo "stopping and removing dp-compose containers and networks"
	docker-compose down

.PHONY: clean
clean:
	@echo "stopping and removing dp-compose containers, associated volumes and networks"
	docker-compose down -v

.PHONY: restart
restart:
	@echo "restarting dp-compose containers"
	docker-compose restart
