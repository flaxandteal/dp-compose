#!/usr/bin/env bash

if test -d "../zebedee"
  then
      # it wll exit 1 but the important bit is it builds the jar file :)
      docker-compose --env-file=start-default.env run --entrypoint "/The-Train/run.sh && exit" the-train
fi
