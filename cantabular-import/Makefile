# start full journey
start:
	sudo -E docker-compose up

# start full journey (containers in background)
.PHONY: start-detached
start-detached:
	@echo "starting containers in background"
	sudo -E docker-compose up -d

# start only backend services
.PHONY: start-backend
start-backend:
	@echo "starting containers"
	sudo -E docker-compose --env-file .env.backend up

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
