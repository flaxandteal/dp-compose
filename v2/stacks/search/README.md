# Search Stack

This stack deploys the necessary services and dependencies for the search functionality.

The search stack uses elasticsearch to store some indexed data that can be queried via the search api.

You may run the stack in stand-alone mode, assuming you already have the data you need in elasticsearch. 

Or you may run it with mappings to localhost, to obtain data from external sources (required if you need to re-index or run the extract-import pipeline with data available externally)


## Run with mappings

If you want to use data from en external source (e.g. Sandbox environment), you may use the backend-with-mappings stack, like so:

1- Set a valid service auth token for the environment you want to use. For example, you may check the environment's secrets in `dp-configs` and use a valid token:

```sh
export export SERVICE_AUTH_TOKEN=<valid_token>
```

2- Gain access to the environment you want to use. For example, you may login to sandbox environment:

```sh
aws sso login --profile dp-sandbox
```

3- Use the `dp` tool to ssh to `zebedee` and `dp-dataset-api` with port forwarding. For example:

```sh
# Zebedee
dp ssh sandbox publishing 1 -p 8082:10.30.138.93:26251
```

```sh
# Dataset API
dp ssh sandbox publishing 2 -p 22000:10.30.138.234:25681
```

Please, replace the publishing node, ip and port according to where the services are currently deployed when you run this. You can check this in [Consul](https://consul.dp.aws.onsdigital.uk/ui/eu/services)

4- Edit docker-compose config

Edit this stack's `.env` file and uncomment the necessary lines to override `ZEBEDEE_URL` and `DATASET_API_URL` with the `host.docker.internal` values.

Uncomment the `COMPOSE_FILE` that uses `backend-with-mappings.yml` and comment the other one.

5- Run the stack

```sh
make start-detached
```

### Reindex

In order to populate elasticsearch, you may run the reindex script, and if you have followed the previous steps you will have access to the necessary external data.

Navigate to your `search-api` location, edit the necessary config under `cmd/reindex/local.go`, and run:

```sh
make reindex
```

For more information on the reindex script, please check [search-api instructions](https://github.com/ONSdigital/dp-search-api/blob/develop/README.md#running-bulk-indexer).

### Extract-Import pipeline

When a dataset is published, the search extract-import kafka pipeline is triggered. You may emulate this in the search stack by using the command line tool to generate kafka messages.

WARNING: The pipeline assumes that an index with alias "ons" already exists, please make sure you have run the re-index script before trying the pipeline.

1- Navigate to your `search-data-extractor` location and then to `cmd/producer`

2- Run `go run main.go` and introduce the requested fields. When all the information is introduced a kafka message will be produced. For example:

```sh
--- [Send Kafka ContentPublished] ---
Please type the URI
$ /datasets/your-datasetid-here/editions/2021/versions/1/metadata
Please type the dataset type (legacy or datasets)
$ datasets
Please type the collection ID
$ collection-id
{"created_at":"2023-03-28T12:49:39.788994Z","namespace":"dp-search-data-extractor","event":"sending content-published event","severity":3,"data":{"contentPublishedEvent":{"URI":"datasets/your-datasetid-here/editions/2021/versions/1/metadata","DataType":"datasets","CollectionID":"collection-id","JobID":"","SearchIndex":"","TraceID":"054435ded"}}}
```

3- Check the docker-compose logs, starting with `search-data-extractor` to validate that the message is consumed and processed as expected.

## Run standalone

If you want to run a stand-alone search stack, without external dependencies, you may use the basic stack, like so:

1- Edit docker-compose config

Edit this stack's `.env` file and comment the necessary lines to prevent `ZEBEDEE_URL` and `DATASET_API_URL` being overwritten.

Uncomment the `COMPOSE_FILE` that uses `backend.yml` and comment the other one.

2- Run the stack

```sh
make start-detached
```

### Run in Publishing mode

To run in publishing mode (mostly used to view Search via Florence) do the following:

1- Edit docker-compose config

Edit this stack's `.env` file and comment the necessary lines to prevent `ZEBEDEE_URL` and `DATASET_API_URL` being overwritten.

Uncomment the `COMPOSE_FILE` that uses `frontend-publishing.yml` and comment the other one.

Uncomment the line that sets `IS_PUBLISHING` to `true`

2- Run the stack

```sh
make start-detached
```
