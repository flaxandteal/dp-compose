#!/usr/bin/env bash 

# To enable trace, uncomment line below
# set -x

##################### VARIABLES ##########################

# prompt colours
GREEN="\e[32m"
RED="\e[31m"
RESET="\e[0m"

# services
SERVICES="babbage florence The-Train zebedee dp-api-router 
dp-cantabular-api-ext dp-cantabular-csv-exporter dp-population-types-api
dp-cantabular-filter-flex-api dp-cantabular-metadata-exporter dp-cantabular-server
dp-cantabular-xlsx-exporter dp-download-service dp-dataset-api dp-filter-api dp-frontend-router 
dp-import-api dp-import-cantabular-dataset dp-import-cantabular-dimension-options dp-recipe-api 
dp-frontend-filter-flex-dataset dp-zebedee-utils dp-cantabular-metadata-service"

EXTRA_SERVICES="dp-cantabular-ui dp-cantabular-server dp-kafka dp-setup dp-cantabular-uat"

# current directory
DIR="$( cd "$( dirname "$0" )/../.." && pwd )"

# directories
DP_COMPOSE_DIR="$DIR/dp-compose"
DP_FLORENCE_DIR="$DIR/florence"
DP_THE_TRAIN_DIR="$DIR/the-train"
DP_CANTABULAR_IMPORT_DIR="$DP_COMPOSE_DIR/cantabular-import"
DP_CANTABULAR_SERVER_DIR="$DIR/dp-cantabular-server"
DP_CANTABULAR_API_EXT_DIR="$DIR/dp-cantabular-api-ext"
DP_FRONTEND_ROUTER_DIR="$DIR/dp-frontend-router"
DP_CANTABULAR_METADATA_SERVICE_DIR="$DIR/dp-cantabular-metadata-service"
DP_FRONTEND_FILTER_FLEX_DATASET_DIR="$DIR/dp-frontend-filter-flex-dataset"
DP_FRONTEND_DATASET_CONTROLLER_DIR="$DIR/dp-frontend-dataset-controller"
DP_RECIPE_API_IMPORT_RECIPES_DIR="$DIR/dp-recipe-api/import-recipes"
DP_DATASET_API_IMPORT_SCRIPT_DIR="$DIR/dp-dataset-api/import-script"
ZEBEDEE_DIR="$DIR/zebedee"
ZEBEDEE_GENERATED_CONTENT_DIR=${zebedee_root}

ACTION=$1

##################### FUNCTIONS ##########################
logSuccess() {
    echo -e "$GREEN ${1} $RESET"
}

logError() {
    echo -e "$RED ${1} $RESET"
}

splash() {
    echo "Start Cantabular Services (SCS)"
    echo ""
    echo "Simple script to run cantabular import service locally and all the dependencies"
    echo ""
    echo "List of commands: "
    echo "   chown     - change the service '.go' folder permissions from root to the user and group."
    echo "               Useful for linux users."
    echo "   clone     - git clone all the required GitHub repos"
    echo "   down      - stop running the containers via docker-compose"
    echo "   init-db   - preparing db services. Run this once"
    echo "   help      - splash screen with all these options"
    echo "   pull      - git pull the latest from your remote repos"
    echo "   setup     - preparing services. Run this once, before 'up'"
    echo "   up        - run the containers via docker-compose"
}

cloneServices() {
    cd $DIR
    allServices="${SERVICES} ${EXTRA_SERVICES}"
    for service in $allServices; do
        git clone git@github.com:ONSdigital/${service}.git 2> /dev/null
        logSuccess "Cloned $service"
    done
    for service in $EXTRA_SERVICES; do
        git clone git@github.com:ONSdigital/${service}.git 2> /dev/null
        logSuccess "Cloned additional $service"
    done
}


pull() {
    cd $DIR
    for repo in $(ls -d $DIR/*/); do
        cd ${repo}
        if [ -d ".git" ]; then
          git pull
          logSuccess "'$repo' updated"
        fi
    done
}


initDB() {
    echo "Importing Recipes & Dataset documents..."
    cd $DP_CANTABULAR_IMPORT_DIR
    make init-db
    if [ $? -ne 0 ]; then
        logError "ERROR - Failed to import MongoDB initial datasets"
        exit -1
    fi
    logSuccess "Importing Recipes & Dataset documents... Done."
}

florenceLoginInfo () {
    logSuccess "Florence is available at http://localhost:8081/florence"
    logSuccess "         if 1st time accessing it, the credentials are: florence@magicroundabout.ons.gov.uk / Doug4l"
}

setupServices () {
    logSuccess "Remove zebedee docker image and container..."
    docker rm -f $(docker ps --filter=name='zebedee' --format="{{.Names}}")
    docker rmi -f $(docker images --format '{{.ID}}' --filter=reference="*zebedee*:*")
    if [ $? -ne 0 ]; then
        logError "ERROR - Docker failed to remove containers and images"
        exit -1
    fi
    logSuccess "Remove zebedee docker image and container... Done."

    logSuccess "Build zebedee..."
    cd $ZEBEDEE_DIR
    git checkout develop
    git reset --hard; git pull
    # mvn clean install
    make build build-reader
    if [ $? -ne 0 ]; then
        logError "ERROR - Failed to build zebedee"
        exit -1
    fi
    logSuccess "Build zebedee...  Done."

    logSuccess "Clean zebedee_root folder..."
    cd $DP_CANTABULAR_IMPORT_DIR
    make full-clean
    if [ $? -ne 0 ]; then
        logError "ERROR - Failed to clean zebedee_root folder"
        exit -1
    fi
    logSuccess "Clean zebedee_root folder... Done."

    logSuccess "Make Assets for dp-frontend-router..."
    cd $DP_FRONTEND_ROUTER_DIR
    make assets
    if [ $? -ne 0 ]; then
        logError "ERROR - Failed to build dp-frontend-router assets"
        exit -1
    fi
    logSuccess "Make Assets for dp-frontend-router... Done."

    logSuccess "Generate prod for $DP_FRONTEND_DATASET_CONTROLLER_DIR..."
    cd $DP_FRONTEND_DATASET_CONTROLLER_DIR
    make generate-prod
    if [ $? -ne 0 ]; then
        logError "ERROR - Failed to generate-prod for 'dp-frontend-dataset-controler'"
        exit -1
    fi
    logSuccess "Generate prod for $DP_FRONTEND_DATASET_CONTROLLER_DIR... Done."

    logSuccess "Generate prod for $DP_FRONTEND_FILTER_FLEX_DATASET_DIR..."
    cd $DP_FRONTEND_FILTER_FLEX_DATASET_DIR
    make generate-prod
    if [ $? -ne 0 ]; then
        logError "ERROR - Failed to generate-prod for 'dp-frontend-filter-flex-dataset'"
        exit -1
    fi
    logSuccess "Generate prod for $DP_FRONTEND_FILTER_FLEX_DATASET_DIR... Done."

    logSuccess "Setup metadata service..."
    cd $DP_CANTABULAR_METADATA_SERVICE_DIR
    make setup
    if [ $? -ne 0 ]; then
        logError "ERROR - Failed to setup 'dp-cantabular-metadata-service'"
        exit -1
    fi
    logSuccess "Setup metadata service... Done."

    logSuccess "Build florence..."
    cd $DP_FLORENCE_DIR
    make node-modules && make generate-go-prod
    if [ $? -ne 0 ]; then
        logError "ERROR - Failed to build 'florence'"
        exit -1
    fi
    logSuccess "Build florence...  Done."

    logSuccess "Build the-train..."
    cd $DP_THE_TRAIN_DIR
    make build
    if [ $? -ne 0 ]; then
        logError "ERROR - Failed to build 'the-train'"
        exit -1
    fi
    logSuccess "Build the-train... Done."

    logSuccess "Preparing dp-cantabular-server..."
    cd $DP_CANTABULAR_SERVER_DIR
    make setup
    if [ $? -ne 0 ]; then
        logError "ERROR - Failed to build 'dp-cantabular-server'"
        exit -1
    fi
    logSuccess "Preparing dp-cantabular-server... Done."

    logSuccess "Preparing dp-cantabular-api-ext..."
    cd $DP_CANTABULAR_API_EXT_DIR
    make setup
    if [ $? -ne 0 ]; then
        logError "ERROR - Failed to build 'dp-cantabular-api-ext'"
        exit -1
    fi
    logSuccess "Preparing dp-cantabular-api-ext... Done."
    
    chown

    upServices

    initDB

    florenceLoginInfo
}

chown() {
    # list of services that have the '.go' folder
    listOfServices=$(ls -l ./*/.go | grep ".go" | awk -F'/' '{print $2}')

    user=$(id -u --name)
    group=$(id -g --name)
    for service in $listOfServices; do
        sudo chown $user:$group -R ${DIR}/${service}/.go
    done

    sudo chown $user:$group -R ${ZEBEDEE_GENERATED_CONTENT_DIR}
    sudo chmod 755 -R ${ZEBEDEE_GENERATED_CONTENT_DIR}
}

upServices () {
    echo "Starting dp cantabular import..."
    cd $DP_CANTABULAR_IMPORT_DIR
    make start-detached
    # make start
    echo "Starting dp cantabular import... Done."
    florenceLoginInfo
}


downServices () {
    echo "Stopping base services..."
    cd $DP_COMPOSE_DIR
    docker-compose down
    logSuccess "Stopping base services... Done."

    echo "Stopping dp cantabular import..."
    cd $DP_CANTABULAR_IMPORT_DIR
    make stop
    logSuccess "Stopping dp cantabular import... Done."
}


#####################    MAIN    #########################

case $ACTION in 
"chown") chown;;
"clone") cloneServices;;
"help") splash;;
"down") downServices;;
"up") upServices;; 
"pull") pull;;
"setup") setupServices;;
"init-db") initDB;;
*) echo "invalid action - [${ACTION}]"; splash;;
esac