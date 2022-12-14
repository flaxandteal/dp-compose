# Stacks

This folder contains definitions for the different stacks. To run a stack you have to go to its corresponding folder:

```shell
cd stacks/<stack_name>
```

### Initialisation:

Please make sure you have cloned the required repositories as siblings of dp-compose.

Some stacks require an initialisation step, please check the corresponding stack instructions in that case.


### Run with default docker compose commands:

Then just standard Docker compose commands: e.g.:

- to start detached:
```shell
docker compose up -d
```

- or with the alias:
```shell
dpc up -d
```

- to get logs for a service:
```shell
docker compose logs -f dp-files-api` or `dpc logs dp-files-api
```

### Run with default docker compose commands:

Alternatevily, if a Makefile is provided for the stack, you can run the corresponding `make` command. For example:

```shell
make start
```

```shell
make clean
```

### Environment

The stacks should run

Check the `.env` file and change it for your development requirements - you might need to point to local services running in an IDE for example.


## Homepage web

Deploys a stack for the home page and census hub in web mode.

1) The first time that you run it, you will need to generate the assets and have the zebedee content. It also assumes htat you have defined `zebedee_root` env var in your system.

2) Go to the stack root folder

```sh
cd homepage-web
```

3) Start the docker stack

```sh
make start-detached
```

3) Open your browser and check you can see the home page: `localhost:24400`

3) Check you can see the census hub page: `localhost:24400/census`

4) You may stop all containers and clean the environment when you finish:

```sh
make clean
```


## Homepage publishing

Deploys a stack for the home page in publishing mode.

1) The first time that you run it, you will need to generate the assets and have the zebedee content. It also assumes htat you have defined `zebedee_root` env var in your system.

2) Go to the stack root folder

```sh
cd homepage-publishing
```

3) Start the docker stack

```sh
make start-detached
```

3) Open your browser and check you can florence website: `http://localhost:8081/florence`

4) Log in to florence and perform a publish journey

5) You may stop all containers and clean the environment when you finish:

```sh
make clean
```


## Interactives

1) Start the docker stack

```sh
make start-detached
```

2) You may stop all containers and clean the environment when you finish:

```sh
make clean
```


## Interactives with auth

1) Start the docker stack

```sh
make start-detached
```

2) You may stop all containers and clean the environment when you finish:

```sh
make clean
```


## Static files

1) Start the docker stack

```sh
make start-detached
```

2) You may stop all containers and clean the environment when you finish:

```sh
make clean
```


## Static files with auth

1) Start the docker stack

```sh
make start-detached
```

2) You may stop all containers and clean the environment when you finish:

```sh
make clean
```
