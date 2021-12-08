# Kafka Cluster

`docker-compose.yml` defines a kafka cluster with 3 kafka nodes and one zookeeper node.

## Topics

Once the cluster is running, you can use `kafka-topics` from your host machine to create or validate topics.

Note that even though `kafka-topics` is available to be run inside the containers, it will fail due to the other nodes in the cluster responding with the `localhost` address instead of the docker network address (`kafka-...` host names).

## Create a topic

You can use `kafka-topics --create` to create a new topic in the cluter.
For example, you can create `myTopic` topic with 60 partitions and 2 replicas with the following command:

```sh
kafka-topics --create --topic myTopic --bootstrap-server kafka-1:19092 --partitions 60 --replication-factor 2
```

## Describe topics

You can use `kafka-topics --describe` to check that a topic has been created.
For example, you can run:

```sh
kafka-topics --describe --bootstrap-server localhost:19092
```

The response will contain a table of partitions for each topic, with the replicas and leader.


## Reference

This compose file has been created using kafka-cluster  example as a reference.
See more  examples https://github.com/confluentinc/kafka-images/tree/master/examples
