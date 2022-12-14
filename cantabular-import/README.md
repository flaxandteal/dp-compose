### Cantabular Import Journey ###

## Setup environment and start services

1. Ensure you have the following environment variable set up on your `~/.zshrc` or `~/.bashrc` profile
```shell
 DP_CLI_CONFIG
 zebedee_root
```
2. Add a new alias to your `~/.zshrc` ou `~/.bashrc` profile:
```shell
alias scs='PATH_TO_ONS_WORKSPACE/dp-compose/cantabular-import/scs.sh'
```
3. Source your profile, or simple close and open a new terminal window
```shell
source ~/.zshrc
```
4. Check if the `scs.sh` is now available on your command line, by simply typing `scs`

   If interested in having an [interactive menu](https://github.com/charmbracelet/gum) please install [**gum**](https://github.com/charmbracelet/gum#installation).
```shell
> scs

Start Cantabular Services (SCS)

Simple script to run cantabular import service locally and all the dependencies

List of commands: 
> [•] chown           change the service '.go' folder permissions from root to the user and group. Useful for linux users.
  [ ] clone           git clone all the required GitHub repos
  [ ] fe-assets       generate Cantabular FE assets
  [ ] init-db         preparing db services. Run this once
  [ ] pull [branch]   by default it will pull the latest from the current branch. Optionally, provide a branch to pull from, e.g., 'scs pull develop'
  [ ] setup           preparing services. Run this once
  [ ] start           run the containers via docker-compose with logs attached to terminal
  [ ] start-detached  run the containers via docker-compose with detached logs (default option)
  [ ] stop            stop running the containers via docker-compose
```

5. Clone all the required GitHub repositories: `scs clone`

6. **Setup your environment and start the servic**e: `scs setup`

   This is intended for when setting up the service for the first time or when a clean setup is required.

   This will:
     * remove `zebedee` container and image and clean `zebedee_root` folder
     * get the latest from `zebedee`, build image
     * get the latest from `dp-frontend-router` and create the static assets
     * generate-prod static assets for `dp-frontend-dataset-controler`
     * generate-prod static assets for `dp-frontend-filter-flex-dataset`
     * setup `dp-cantabular-metadata-service`, `dp-cantabular-server` and `dp-cantabular-api-ext`
     * build `florence` and `the-train`
     * start the docker microservices via [docker-compose](https://docs.docker.com/engine/reference/commandline/compose/)
     * seed the MongoDB collections. If this fails please run `scs init-db`
     * provide florence login details for first-time user


7. To Start the local environment: `scs start-detached`
8. To stop the local environment: `scs down`


## Debugging 

* Go services will automatically rebuild upon detecting source file changes
* If you need to make adjustments to compose files etc, you can just run `scs start-detached` and docker-compose will
  automatically detect which services need rebuilding (no need to bring everything down first).
* On Mac/Darwin the memory resources allocated to docker desktop may not be sufficient
  If microservices stop working unexpectedly, please provide more resources to it
* `scs start-detached` will try 3x to bring up all the required services. If after that, not all cantabular services are up an running it will print a warning message to the console.
* Sometimes seeding the DB may occur before the MongoDB is ready to accept new operations. Please run `scs init-db` if no recipes have been imported
* Rebuild `dp-cantabular-server` and `dp-cantabular-api-ext` docker images whenever there are new code updates
* To allow [SSH remote port forwarding](https://github.com/ONSdigital/dp-cli#ssh-commands) you can use ``dp ssh [environment] [subnet] [host] [port]``.

  _For example, to allow port forwarding to sandbox's publishing subnet host **1** on port 14500:_ ``dp ssh sandbox publishing 1 -p 10450:10450`` 

## Frontend Setup of FE Developers

There is a CORS error when loading the `dp-design-system` JavaScript (JS) files resulting in JS interactivity to fail.

To test JS interactions, you can set/add the `debug` environment variable to `true` in `dp-frontend-filter-flex-dataset` 
and `dp-frontend-dataset-controller` then run the [dp-design-system](https://github.com/ONSdigital/dp-design-system#readme).

### Local FE Setup
* In `dp-compose/cantabular-import/dep.yml` - remove line 92 (-`'9002:9001'`) := prevents port binding conflict with dp-design-system
* All repos on head of develop branch
* `dp-compose/cantabular-import docker compose --env-file .env.backend up -d`
* For **dp-api-router**: `make debug ENABLE_PRIVATE_ENDPOINTS=true ENABLE_POPULATION_TYPES_API=true`
* For **dp-frontend-router**: `make debug FILTER_FLEX_ROUTES_ENABLED=true`
* For **dp-frontend-dataset-controller**:  `make debug ENABLE_CENSUS_PAGES=true`
* For **dp-frontend-filter-flex-dataset**: `make debug`
* For **dp-design-system**: `npm run dev`
* For **florence** _(only if actively developing within)_: `make node-modules && make debug`
* Use **postman** to get the `X-Florence-Token` and insert value into browser cookie `access_token`

## Alternative to start the Cantabular Import Services

This assumes that all your setup has been done and all you need to do 
is to start the docker network, but may not want to use the `scs` utility script.
Ensure you are in the `dp-compose/cantabular-import` folder.

* **Start all Services**: `make start`
* **Start services in the background**: `make start-detached`
* **Stop Services**: `make stop`
* **Stop Services And Remove Containers**: `make down`
* **Stop Services And Remove All Containers, Volumes and Networks**: `make clean`
* **Restart Services**: `make restart`
* **Recall Logs**: `make logs` or `./logs`
* **Recall Logs For Specific Service**: `make logs t=<service-name>` or `./logs <service-name>`

------------------
Files:

    run-cantabular-without-sudo.sh

    get-florence-token.sh

are used by `cantabular-import/helpers/test-compose/test-compose.go` and are nedded at this level for it to bring up the cantabular containers.

------------------

# Adding New Services To Journey #

If you need to add a new service to the journey you need to take the following steps.

For a Golang service:

- Add `Dockerfile.local` and `reflex` file to service root directory. Copy from existing
examples and change instances of the service name in Dockerfile.local and the command
to be executed. This usually either `make debug` or `make debug-run`. Check the Makefile
and use the make target that runs the service using `go run` (as opposed to building an
executable and then running it). If there isn't such a target you will need to add it.
Again, use existing examples as a guide.
- Add `/go` to .`dockerignore` for service.
- Add `dp-my-service-name.yml` to `dp-compose/cantabular-import` directory. Follow existing
examples as a guide. Be sure to use the correct service name, port and environment variables
the service will need. These include those used when you would usually run the service (e.g.
`ENABLE_PRIVATE_ENDPOINTS=true`) and others that would usually use the service's default.
Most commonly these include URLs to other services which will need to be set to the
`http://service-name:port` from `http://localhost:port`. Also sometimes the `BIND_ADDR` will
default to `http://localhost:port` which will need to be set to simply `:port`.
- Add service yml file to relavent `.env` files in `dp-compose/cantabular-import`. The default
`.env` is for the full journey including front-end services, `env.backend` for back-end only
services and new `.env` files can be created for different journeys with different collections
of services as needed.
- If there are any external services (e.g. MongoDB, Kafka) the service depends on that are
not already included in the compose cluster add them to `deps.yml`. Be cognizant of which
services the new service is dependant on and add under the `depends_on` clause.
- Test the new service runs as expected and can be reached by other services and you're
good to go!


# Monitoring Apache Kafka clusters #

On the Docker network we have a Kowl as our UI to monitor our Kafka Cluster on the docker network.
The tools display information such as brokers, topics, partitions, consumers, message schemas and 
lets you view messages on our docker network.

* [**Kowl**](https://github.com/redpanda-data/kowl) - [http://localhost:9888](http://localhost:9090)

# Developing services and libraries with the stack

While developing service code any changes will be automatically picked up by `reflex` when files are
saved and the services will be automatically rebuilt.

When developing a library that is used by the services as a dependency (i.e. `dp-api-clients-go`,
`dp-net` etc) a couple of extra steps are needed to pick up local changes. Firstly, in order to
point the service to use your local copy of the library you will need to add a replace directive
in `go.mod` pointing to the local copy you're working on. For example

`replace github.com/ONSdigital/dp-api-clients-go/v2 => LOCAL_PATH_TO/dp-api-clients-go`

Secondly you'll need to add a volume to the service's docker container in dp-compose. For example to
work with the above in `dp-filter-api` you would add:

`- LOCAL_PATH_TO/dp-api-clients-go:LOCAL_PATH_TO/dp-api-clients-go`

under `volumes:` in `dp-compose/cantabular-import/dp-filter-api.yml`

To pick up changes made to the library you need to save/update a file in the service that imports the
library. Reflex won't automatically detect changes made to the library code itself.
