### Cantabular Import Journey ###

## Requirements ##

Expects you to have environment variables `zebedee_root` and 
`SERVICE_AUTH_TOKEN` set in your local environment

Expects your services to be in the expected relative path

# Bring Up Cantabular Import Services #

`sudo docker-compose up`

# Bring Up Cantabular Import Services Detached (running in background) #

`sudo docker-compose up -d`

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


