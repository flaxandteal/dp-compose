### Helper scripts ###

Find here a some scripts to help with working with the Cantabular import process

## florence-token ##

`florence-token` returns a fresh florence token to use with requests that require
it.

Usage: edit `florence-token` to have your username/password and run
`./florence-token`

## start-import.sh ##

`start-import.sh` starts an import of the Example Cantabular Dataset.

Usage: 
* Make sure you have your Cantabular Import Journey set as per
`cantabular-import/README.md`

* Edit the `./florence-token` script to use your florence username/password

* Run `./start-import.sh`

Alternatively if you have your own script/other source to get a florence token
you can pipe the output directly into `go run start-import/main.go`

for example, if you had a FLORENCE_TOKEN environment variable saved you could run
`echo $FLORENCE_TOKEN | go run start-import/main.go`

-----

## Validating integration process utils ##

In order to run the 'start-import.sh` script which kicks off one import job, you need to have the docker files for Cantabular running. 

Sometimes a verion of docker / docker-compose may not prove reliable, so the `test-compose.sh` script can be used to gain confidence in docker and docker-compose.

The script `test-compose.sh` runs multiple iterations of these steps:

    start containers
    get florence token
    run import process
    stop containers.

Its primary purpose is to determine that the version of Docker and Docker-compose are working well (it also demonstrates that the whole import process is functioning, as are all the contaiers / services)

Usage: (on mac books)
* Have a version of Docker running that you wish to test.

    Using Docker 3.3.3 works well (the one before having docker-compose v2.0.0-beta.6 ... which does not work well)

* Edit the `../get-florence-token.sh` script to use your florence username/password

* Adjust the constant `maxRuns` in `test-compose.go` for the number of times you want the process to run. Each loop of the process may take about 3 minutes.

* Run `./test-compose.sh`

* follow any instructions the above produces.

## Examining docker logs ##

After running one or more import processes, you can extract the logs for the containers by doing:

run:

    ../run-cantabular-without-sudo.sh

Then wait for all the containers to be running, pause for another 15 seconds to be sure, then run:

    ./extract-docker-logs.sh

You can then run two more utilities to get useful information about the import process(es) run by running the following:

    ./count-log-events.sh

    ./extract-job-info.sh

The information that these utils produce as an output in the terminal (and saved within their directories) can help to determine if the whole import process is working as desired.
