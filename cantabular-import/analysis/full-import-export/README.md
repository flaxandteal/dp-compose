To use this application, follow these steps:

STEP 1:

In one terminal, goto to directory:

src/github.com/ONSdigital/dp-compose/cantabular-import

and run:

./run-cantabular-without-sudo.sh

STEP 2:

In another terminal, goto to directory:

src/github.com/ONSdigital/dp-compose/cantabular-import

and run:

make init-db

STEP 3:

In another terminal run:

docker stats

and wait until the CPU%'s for all the apps has fallen to close to zero, kafka may stay above 10% and below 30%

STEP 4:

In another terminal, goto to directory:

src/github.com/ONSdigital/dp-compose/cantabular-import/analysis

and run:

./full-import-export.sh

This should run a full import and export.

If there are any errors you may need to in the terminal for STEP 1, do CTRL-C and when the docker compose finishes (which it may not), run 'docker compose down' (this may also fail - in which case restart docker and do 'docker compose down' again) - after that clean out any volumes in docker and repeat all 4 steps again. Hopefully STEP 4 will then run to completion and in your local minio directory you should find 4 newly created files in the private bucket and also 4 files in public bucket.
