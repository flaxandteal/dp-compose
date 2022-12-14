#!/usr/bin/env bash 

# To enable trace, uncomment line below
# set -x

##################### VARIABLES ##########################

# prompt colours
GREEN="\e[32m"
RED="\e[31m"
RESET="\e[0m"

# services
SERVICES="zebedee dp-topic-api dp-api-router dp-frontend-router dp-frontend-homepage-controller"

# current directory
DIR="$( cd "$( dirname "$0" )/../../../.." && pwd )"

# directories
DP_FRONTEND_ROUTER_DIR="$DIR/dp-frontend-router"
DP_FRONTEND_HOMEPAGE_CONTROLLER_DIR="$DIR/dp-frontend-homepage-controller"
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
    echo "Start Homepwage web Services"
    echo ""
    echo "Simple script to run homepage web stack service locally and all the dependencies"
    echo ""
    echo "List of commands: "
    echo "   clone     - git clone all the required GitHub repos"
    echo "   down      - stop running the containers via docker-compose"
    echo "   init-db   - preparing db services. Run this once"
    echo "   help      - splash screen with all these options"
    echo "   pull      - git pull the latest from your remote repos"
    echo "   setup     - preparing services. Run this once, before 'up'"
    echo "   up        - run the containers via docker-compose"
}

cloneServices() {
    echo "Repositories to clone:"
    for service in ${SERVICES}; do
        echo " - ${service}"
    done
    read -p "Repos will be cloned to $DIR, continue (y/n)?" -n 1 -r
    echo    # new line
    if [[ ! $REPLY =~ ^[Yy]$ ]]
    then
        [[ "$0" = "$BASH_SOURCE" ]] && exit 1 || return 1 # handle exits from shell or function but don't exit interactive shell
    fi

    cd "$DIR"
    for service in ${SERVICES}; do
        git clone git@github.com:ONSdigital/${service}.git 2> /dev/null
        logSuccess "Cloned $service"
    done
}

pull() {
    echo "Repositories to pull:"
    for service in ${SERVICES}; do
        echo " - ${service}"
    done
    read -p "Assuming root folder is $DIR, continue (y/n)?" -n 1 -r
    echo    # new line
    if [[ ! $REPLY =~ ^[Yy]$ ]]
    then
        [[ "$0" = "$BASH_SOURCE" ]] && exit 1 || return 1 # handle exits from shell or function but don't exit interactive shell
    fi

    for service in ${SERVICES}; do
        cd "$DIR/${service}"
        git pull
        logSuccess "$service updated"
    done
}

initDB() {
    echo "initDB not implemented"
}

florenceLoginInfo () {
    logSuccess "Florence is available at http://localhost:8081/florence"
    logSuccess "         if 1st time accessing it, the credentials are: florence@magicroundabout.ons.gov.uk / Doug4l"
}

checkEnvironmentVariables() {
  if [ -z $(echo "$SERVICE_AUTH_TOKEN") ]; then
    logError "Error - Environment variable [SERVICE_AUTH_TOKEN] is not defined"
    exit 129
  fi

  if [ -z $(echo "$zebedee_root") ]; then
    logError "Error - Environment variable [zebedee_root] is not defined"
    exit 129
  fi
}

setupServices () {
    checkEnvironmentVariables

    logSuccess "Make Assets for dp-frontend-router..."
    cd "$DP_FRONTEND_ROUTER_DIR"
    git checkout develop && git pull
    make assets
    if [ $? -ne 0 ]; then
        logError "ERROR - Failed to build dp-frontend-router assets"
        exit 129
    fi
    logSuccess "Make Assets for dp-frontend-router... Done."

    logSuccess "Make Generate Debug for dp-frontend-homepage-controller..."
    cd "$DP_FRONTEND_HOMEPAGE_CONTROLLER_DIR"
    git checkout develop && git pull
    make generate-debug
    if [ $? -ne 0 ]; then
        logError "ERROR - Failed to build dp-frontend-homepage-controller assets"
        exit 129
    fi
    logSuccess "Make Generate Debug for dp-frontend-router... Done."
}

upServices () {
    echo "Starting homepage stack in web mode..."
    make start-detached
    # make start
    echo "Starting homepage stack in web mode... Done."
    # florenceLoginInfo
}


downServices () {
    echo "Stopping homepage stack web services..."
    docker-compose down
    logSuccess "Stopping homepage stack web services... Done."
}


#####################    MAIN    #########################

case $ACTION in 
"clone") cloneServices;;
"help") splash;;
"down") downServices;;
"up") upServices;; 
"pull") pull;;
"setup") setupServices;;
"init-db") initDB;;
*) echo "invalid action - [${ACTION}]"; splash;;
esac
