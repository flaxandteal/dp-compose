# start full journey
.PHONY: start
start:
	UID=$(id -u) GID=$(id -g) docker-compose up

# Start full journey with super user privileges
.PHONY: start-privileged
start-privileged:
	sudo docker-compose up

# start full journey (containers in background)
.PHONY: start-detached
start-detached:
	@echo "starting containers in background"
	UID=$(id -u) GID=$(id -g)  docker-compose up -d

# start only backend services
.PHONY: start-backend
start-backend:
	@echo "starting back-end containers"
	UID=$(id -u) GID=$(id -g)  docker-compose --env-file .env.backend up

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

.PHONY: full-clean
full-clean: clean
	@echo recreating zebedee content
	cd helpers && ./clean-zebedee.sh

.PHONY: restart
restart:
	@echo "restarting containers"
	docker-compose restart

.PHONY: init-db
init-db:
	@echo "initialising Mongo with recipes and CPIH dataset"
	cd helpers && ./init-db.sh

# print logs for given service
.PHONY: logs
logs:
	./logs $t
