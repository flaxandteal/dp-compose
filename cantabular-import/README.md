### Cantabular Import Journey ###

## Requirements ##

Expects you to have environment variables `zebedee_root` and 
`SERVICE_AUTH_TOKEN` set in your local environment

Expects your services to be in the expected relative path

You will need to run the `import-recipes` script in `dp-recipe-api` when
first building the containers before running an import.

# Bring Up Cantabular Import Services #

`sudo -E docker-compose up` or `./run.sh`

# Bring Up Cantabular Import Services Detached (running in background) #

`sudo -E docker-compose up -d`

# Stop Services #

`docker-compose down`

# Recall Logs For Specific Service #

`docker-compose logs -f <service-name>` or `./logs <service-name>`

## Notes ##

# Making Changes #

Go services will automatically rebuild upon detecting source file changes.

If you need to make adjustments to compose files etc, you can just
run `docker-compose up -d` and docker-compose will automatically detect 
which services need rebuilding (no need to bring everything down first).

## Known Issues ##

dp-dataset-api requires a connection to a graph db in order to not complain
at healthcheck time. As it stands the service isn't configured to be able
to access the ssh tunnel to Neptune from the container's host machine.

This doesn't prevent the Cantabular import journey from working but it does
produce an error log.

Either dp-dataset-api needs to be able to be configured to not require a
graph db connection or the container needs to be configured to access the
host environments localhost.

