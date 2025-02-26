# dp-compose
A project to assist in composing multiple DP services

Running dp-compose assumes Docker is running natively and not in a VM. On a Mac this requires Docker for mac - NOT the previous docker toolbox which runs docker in a VM.

Note that if you run Docker using the mac VM, you will need to increase its resources.

https://www.docker.com/products/docker#/mac

More information about the kafka cluster [here](./kafka-cluster.md)

### V2

There is a version 2 folder, which contains different docker compose definitions for each stack.

Please, have a look to [version 2 README](./v2/README.md) for more information

### Run

You may run containers for all required backing services by doing one of the following:
- Run ```docker-compose up```
- Using the ``` ./run.sh ``` script does the same thing.
- Run `make start` to start the kafka cluster containers

You can run `make stop` to stop the containers, or `make clean` to stop and remove them as well.

**If you're running for the first time** you will need to seed the Mongo database and create the collections for the first time. Run the [init-db.sh](https://github.com/ONSdigital/dp-compose/blob/main/cantabular-import/helpers/init-db.sh) script to create recipe and dataset related collections.

## CMD

The ONS website and CMD both require Elastic search but (annoyingly) require different versions. The `docker-compose.yml` will start 2 instances. 

**Note:** The default ports for Elastic search is usually `9200` & `9300` however in order to avoid a port conflict
 when running 2 different versions on the same box at the same time the CMD instance is set to use ports `10200` & `10300`.

:warning: **Gotcha Warning** :warning:
You'll need to overwrite your ES config for the `dp-dimension-search-builder` and `dp-dimension-search-api` to use ports `10200` & `10300` to ensure they are using the correct instance.

## Postgis

**Important**: Zebedee requires _**Postgis**_. 

The _dp-compose_postgis_ container by default uses port `5432` 

### Checking postgres version

`docker ps -a`

You should see something similar to (see IMAGE):
```
CONTAINER ID    IMAGE                 COMMAND                   CREATED           STATUS           PORTS                     NAMES
d343558fd467    dp-compose_postgis    "docker-entrypoint.s…"    11 minutes ago    Up 11 minutes    0.0.0.0:5432->5432/tcp    dp-compose-postgis-1
```

Or (see TAG)
```
docker images
```
```
REPOSITORY    TAG     IMAGE ID        CREATED           SIZE
postgis      latest   ed34a2d5eb79    25 minutes ago    567MB
```

### Connecting to Postgres
To connect to the container and query via the postgres _cli_

```
docker run -it --rm --link dp-compose_postgis_1:postgis --net dp-compose_default postgis/postgis psql -h postgis -U postgres
```

## Versioning

Dependencies should be kept at specific versions and up-to-date with production.
Previously we were just using 'latest' and out-of-date versions which both could lead to unexpected behaviour.
This repository should be the source of truth for which versions to use for dependencies. 
