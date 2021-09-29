#!/bin/bash

cd ../../../dp-recipe-api/import-recipes && ./import-recipes.sh mongodb://localhost:27017
cd ../../dp-dataset-api/import-script && ./import-script.sh
