To use this application, follow these steps:

Ensure the florence password is configured in these files:

cantabular-import/get-florence-token.sh
cantabular-import/helpers/get-florence-token.sh


STEP 1:

In one terminal, goto to directory:

src/github.com/ONSdigital/dp-compose/cantabular-import

and run:

make start

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

and in the 'analysis' directory run:

./full-import-export.sh

This should run a full import and export.

If there are any errors you may need to in the terminal for STEP 1, do CTRL-C and when the docker compose finishes (which it may not), run 'make clean' (this may also fail - in which case restart docker and do 'make clean' again) - after that clean out any volumes in docker and repeat all 4 steps again. Hopefully STEP 4 will then run to completion and in your local minio directory you should find 4 newly created files in the private bucket and also 4 files in public bucket.

-=-

If all goes well, you will eventually see the final message:

"ALL steps completed OK ..."


This will also have added the time and id of this process into file:

tmp/id.txt

for later use.

-=-

STEP 5:

After you have done the above steps and the docker containers are still running do:

From the analysis directory do:

./extract-docker-logs.sh

which will extract all the logs for all the containers into:

tmp/all-container-logs.txt

for you to examine.

-=-

If you keep the docker containers running you can repeat step 4 and 5 to add more and more logs
to tmp/all-container-logs.txt


-=-=-

STEP 6:

If you wish to easily examine the logs for just one run of step 4, do the following:

a. First ensure the file tmp/all-container-logs.txt is empty
b. run step 4 again
c. run step 5 again
d. run:
  ./count-log-events.sh

  This will extract and sort by time all the logs that start with the first apperance of the id
  from the file tmp/all-container-logs.txt upto the last log containing that id and save the
  results into file count-log-events/instance-events.txt

  NOTE: As the result file is sorted by time, to shorten the length of the log lines for
        visual inspection, the created at timestamp has been removed.

  This allow you to focus on just the logs from one run of the import/export process.

-=-=-

STEP 7:

All the steps in step 6 can alternatively be achieved by using:

./test-and-logs-by-id.sh

This also a python script that checks for errors for the import/export job run,
and checks for DATA RACE in all logs for all containers since they were started.

-=-=-

the system tesh script and also take two parameters thus:

./test-and-logs-by-id.sh 1 skip

which runs the test once and skips cjecking if all the containers have started, which is a situation
that you might deliberately want if you are doing some work with healthcheckers and for example
you only want 2 of 3 kafka containers running.

or you might run it as:

./test-and-logs-by-id.sh 15

to run the system test 15 times ... which is typically what is needed to observe DATA RACE's
as they don't always show up on just one run.

-=-=-
