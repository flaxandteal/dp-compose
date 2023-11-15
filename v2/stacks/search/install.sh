#!/usr/bin/env bash 

. ../../scripts/utils.sh

# current directory
DIR="$( cd "$( dirname "$0" )/../../../.." && pwd )"

CORE="babbage dp-dataset-api dp-frontend-router zebedee dp-api-router dp-design-system sixteens"
SEARCH="dp-search-api dp-frontend-search-controller dp-search-data-importer dp-search-data-extractor"
RELEASE_CALENDAR="dp-release-calendar-api dp-frontend-release-calendar"
REINDEX="dp-search-reindex-api dp-search-data-finder dp-search-reindex-tracker"
SERVICES="$CORE $SEARCH $RELEASE_CALENDAR $REINDEX"
echo "Installing into $DIR"

cloneServices

echo "Finished installing repositories for search service"
