# Profiles

This folder contains definitions for the different stacks. To run a stack:

1) Go to the folder for the stack you want to run:

```shell
cd profiles/<stack_name>
```

2) Start the services via docker compose

```sh
docker compse up -d
```

Alternatevily, if a Makefile is provided for the stack, you can run the corresponding `make` command. For example:

```shell
make start
```

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

## Homepage publishing

Deploys a stack for the home page in publishing mode

TODO: implement
