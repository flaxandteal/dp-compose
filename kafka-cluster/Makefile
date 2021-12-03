# start kafka cluster
.PHONY: start
start:
	@echo "starting kafka cluster containers"
	docker-compose up

# start kafka cluster (containers in background)
.PHONY: start-detached
start-detached:
	@echo "starting kafka cluster containers in background"
	docker-compose up -d

# stop kafka cluster
.PHONY: stop
stop:
	@echo "stopping kafka cluster containers"
	docker-compose stop

# bring kafka cluster down
.PHONY: down
down:
	@echo "stopping and removing kafka cluster containers and networks"
	docker-compose down

# bring kafka cluster containers down, including volumes
.PHONY: clean
clean:
	@echo "stopping and removing kafka cluster containers, associated volumes and networks"
	docker-compose down -v

# restart kafka cluster containers
.PHONY: restart
restart:
	@echo "restarting kafka cluster containers"
	docker-compose restart

# create topic with 60 partitions and 3 replicas
.PHONY: init-topics
init-topics:
	@echo "initialising Kafka cluster with topics"
	docker exec -it kafka-cluster_kafka-1_1 kafka-topics --create --topic myTopic --bootstrap-server kafka-1:19092 --partitions 60 --replication-factor 3

# describe topics
.PHONY: describe-topics
describe-topics:
	docker exec -it kafka-cluster_kafka-1_1 kafka-topics --describe --bootstrap-server kafka-1:19092
