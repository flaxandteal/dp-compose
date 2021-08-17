## Examining docker logs ##

After running one or more import processes, such as:

`./start-analysis.sh` - for on import
`./parallel-10.sh` - for 10 imports launched in parallel

you can extract the logs for the containers by doing:

* Edit the `../get-florence-token.sh` script to use your florence username/password

run:

    ../run-cantabular-without-sudo.sh

Then wait for all the containers to be running, pause for another 15 seconds to be sure, then run:

    ./extract-docker-logs.sh

You can then run two more utilities to get useful information about the import process(es) run by running the following:

    ./count-log-events.sh

    ./extract-job-info.sh

The information that these utils produce as an output in the terminal (and saved within their directories) can help to determine if the whole import process is working as desired.

### Further points when repeating tests and doing different sorts of tests as per your needs:

1. When jobs are run, their id is placed in / appended to file tmp/id.txt that allows quick access if you have the whole of the dp-compose directory open in say vscode or goland ide's.
2. If you wish to only examine one import process at a time then clear out the contents of the id.txt file before a run of the import process. You will also need to run `docker-compose down` to clear out Docker's logs before a new run.
3. The script `./parallel-10.sh` is an example of kicking off 10 jobs to run in parallel, and when run (needs same pre-requsites as per start-import.sh) you will see 10 id's in the id.txt file. Then running the `extract-docker-logs.sh` followed by either or both of: `count-log-events.sh`, `extract-job-info.sh` you can check if you go the exepeted results for 10 jobs.
4. The files provided for examining the logs perform the actions they do and may serve as a starting point for you to create new apps for different kinds of analysis or for you to adjust locally on your machines to suite the nature of a particular test you wish to run ...