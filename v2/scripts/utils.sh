#!/usr/bin/env bash 

# prompt colours
GREEN="\e[32m"
YELLOW="\e[33m"
RED="\e[31m"
RESET="\e[0m"

# For cloneServices, please set:
#  - $SERVICES - a space separated list of services
#  - $DIR - the directory to install services
cloneServices() {

    cd "$DIR"

    for service in $SERVICES; do
        git clone git@github.com:ONSdigital/${service}.git 2> /dev/null
        logSuccess "Cloned $service"
    done
}

logSuccess() {
    echo -e "$GREEN${1}$RESET"
}

logWarning() {
    echo -e "$YELLOW${1}$RESET"
}

logError() {
    echo -e "$RED${1}$RESET"
}

