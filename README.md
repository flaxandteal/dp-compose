# dp-compose
A project to assist in composing multiple DP services

Running dp-compose assumes Docker is running natively and not in a VM. On a Mac this requires Docker for mac - NOT the previous docker toolbox which runs docker in a VM.

https://www.docker.com/products/docker#/mac

Run ```docker-compose up``` to create docker containers for all required backing services. Using the ``` ./run.sh ``` script does the same thing.


## CMD
The ONS website and CMD both require Elastic search but (annoyingly) require different versions. The `docker-compose.yml` will start 2 instances. 

**Note:** The default ports for Elastic search is usually `9200` & `9300` however in order to avoid a port conflict
 when running 2 different versions on the same box at the same time the CMD instance is set to use ports `10200` & `10300`.

:warning: **Gotcha Warning** :warning: You'll need to overwrite your ES config for the `dp-search-builder` and `dp
-search-api` to use ports `10200` & `10300` to ensure sure its using the correct instance.

## Postgres

**Important**: Zebedee requires _**Postgres 9.6**_. 
If you have error when Zebedee starts up due to a failed database connection make sure that your postgres images is version **9.6*

The _dp-compose_postgres_ container by default uses port `5432` 

### Checking postgres version

`docker ps -a`

You should see something similar to (see IMAGE):
```
CONTAINER ID    IMAGE           COMMAND                   CREATED           STATUS           PORTS                     NAMES
d343558fd467    postgres:9.6    "docker-entrypoint.s…"    11 minutes ago    Up 11 minutes    0.0.0.0:5432->5432/tcp    dp-compose_postgres_1
```

Or (see TAG)
```
docker images
```
```
REPOSITORY    TAG    IMAGE ID        CREATED       SIZE
postgres      9.6    ed34a2d5eb79    3 days ago    230MB
```

If you have a newer version of postgres you can remove it:

```
docker rmi <IMAGE_ID>
```

### Connecting to Postgres
To connect to the container and query via the postgres _cli_

```
docker run -it --rm --link dp-compose_postgres_1:postgres --net dp-compose_default postgres:9.6 psql -h postgres -U postgres
```
