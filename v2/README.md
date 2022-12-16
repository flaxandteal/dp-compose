# v2

Motivation is to consolidate: 
- dp-compose (this repo)
- https://github.com/ONSdigital/dp-static-files-compose
- https://github.com/ONSdigital/dp-interactives-compose

And to create a structure that allows stacks to be easily modified or created.

Giving an end-to-end development environment for working with ONS services in a stable, reliable and repeatable way.

Eventually move this v2 directory as root - remove all other directories/files. So a single source of truth.

## Setup

Completely optional but it might be a good idea to clean the Docker environment - purge all containers/volumes/images and start fresh. Any issues then definitely give this a go first.

For everything to work as expected make sure of the following:

- `git clone` in same dir and level all relevant ONS repos in [manifests](manifests) dir (you will see errors if this is not as expected)
- optional: add an alias - something like `alias dpc='docker-compose --project-dir . -f profiles/static-files.yml'` this makes life a bit easier

## Usage

Please follow the instructions in [stacks README](./stacks/README.md) to run each stack

## Code structure

The required configs and scripts have been structured as follows:

### dockerfiles

Contains `Dockerfile.dp-compose` files for services that do not have a `Dockerfile.local` yet. Each repository should have its own `Dockerfile.local`, so this `dockerfiles` folder can be removed when this is the case.

### manifests

Contains docker compose config `yml` files for each service that is required by any of the stacks. These configurations are stack agnostic and define all the necessary env vars to run the services in any possible configuration that might be required by any stack. Each env var has a sensible default value, which will be used if not provided by the stack, and usually corresponds to the default value in the service config.

The files are organised in subfolders according to their type:
- core-ons: Core services implemented by ONS
- deps: Dependencies, not implemented by ONS, used by ONS services

### stacks

Contains definitions for each stack, including config overrides and docker compose extension files.
Each subfolder corresponds to a particular stack and contains at least:
- {stack}.yml: Extended docker-compose file which uses the manifests for required services.
  - More information [here](https://docs.docker.com/compose/extends/)
- .env: With the environmental variables required to override the default config for the services in the stack
  - More information [here](https://docs.docker.com/compose/environment-variables/#using-the---env-file--option) and [also here](https://docs.docker.com/compose/environment-variables/#using-the---env-file--option)

### provisioning

Contains scripts and files to set the initial state required for stacks to work. This include things like database collections, content, etc.

## Kafka

This uses KRaft'mode but this is early release: https://github.com/apache/kafka/blob/6d1d68617ecd023b787f54aafc24a4232663428d/config/kraft/README.md - have followed this issue but am documenting in case things start to fail: https://github.com/bitnami/bitnami-docker-kafka/issues/159
