### Cantabular Import Journey ###

## Requirements ##

Make sure you have the following repositories cloned to the same root directory
as `dp-compose` (this repository):

`dp-cantabular-server`
`dp-dataset-api`
`dp-import-api`
`dp-import-cantabular-dataset`
`dp-import-cantabular-dimension-options`
`dp-recipe-api`
`zebedee`

Expects you to have environment variables `zebedee_root` and 
`SERVICE_AUTH_TOKEN` set in your local environment

You will need to run the `import-recipes` script in `dp-recipe-api` when
first building the containers before running an import.

# Bring Up Cantabular Import Services #

`make start`

(note: we use `sudo` to prevent docker having issues accessing the `GOCACHE`
volume it creates. `sudo` requires the `-E` in order to preserve existing
environment variables)

# Bring Up Cantabular Import Services Detached (running in background) #

`make start-detached`

# Stop Services #

`make stop`

# Stop Services And Remove Containers #

`make down`

# Stop Services And Remove All Containers, Volumes and Networks #

`make clean`

# Restart Services #

`make restart`

# Recall Logs #

`make logs` or `./logs`

# Recall Logs For Specific Service #

`make logs t=<service-name>` or `./logs <service-name>`

## Notes ##

# Making Changes #

Go services will automatically rebuild upon detecting source file changes.

If you need to make adjustments to compose files etc, you can just
run `make start-detached` and docker-compose will automatically detect 
which services need rebuilding (no need to bring everything down first).
