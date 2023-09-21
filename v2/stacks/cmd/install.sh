#!/usr/bin/env bash 

. ../../scripts/utils.sh

# current directory
DIR="$( cd "$( dirname "$0" )/../../../.." && pwd )"

CORE="babbage sixteens dp-frontend-router dp-api-router zebedee florence dp-frontend-dataset-controller"
IMPORT="dp-dataset-api dp-import-tracker dp-recipe-api dp-hierarchy-api dp-hierarchy-builder dp-dimension-extractor dp-dimension-importer dp-observation-extractor dp-observation-importer dp-upload-service dp-import-api dp-publishing-dataset-controller dp-dimension-search-builder dp-code-list-api"
SERVICES="$CORE $IMPORT"
echo "Installing into $DIR"

cloneServices
