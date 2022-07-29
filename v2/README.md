# v2

Motivation is to consolidate: 
- dp-compose (this repo)
- https://github.com/ONSdigital/dp-static-files-compose
- https://github.com/ONSdigital/dp-interactives-compose

Giving an end-to-end development environment for working with ONS services in a stable, reliable and repeatable way.

Eventually move this v2 directory as root - remove all other directories/files. So a single source of truth.

## docker-compose

This is basically a docker-compose stack run via:
- an env-file: https://docs.docker.com/compose/environment-variables/#using-the---env-file--option also https://docs.docker.com/compose/reference/envvars/
- an extended docker-compose file: https://docs.docker.com/compose/extends/

## Setup

Completely optional but it might be a good idea to clean the Docker environment - purge all containers/volumes/images and start fresh. Any issues then definitely give this a go first.

For everything to work as expected make sure of the following:

- `git clone` in same dir and level all relevant ONS repos in [manifests](manifests) dir (you will see errors if this is not as expected)
- optional: add an alias - something like `alias dpc='docker-compose -f docker-compose.static-files.yml'` this makes life a bit easier

## Usage

Edit `.env` for your development requirements - you might need to point to local services running in an IDE for example.

Then just standard Docker compose commands: e.g.:
- to start detached: `docker-compose -f docker-compose.static-files.yml up -d`
- or with the alias: `dpc up -d`
- to get logs for a service: `docker-compose -f docker-compose.static-files.yml logs dp-files-api`

## Kafka

This uses KRaft'mode but this is early release: https://github.com/apache/kafka/blob/6d1d68617ecd023b787f54aafc24a4232663428d/config/kraft/README.md - have followed this issue but am documenting in case things start to fail: https://github.com/bitnami/bitnami-docker-kafka/issues/159