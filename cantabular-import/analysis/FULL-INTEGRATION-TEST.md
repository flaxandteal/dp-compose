Open 4 terminals:
TL (top left)    : cd dp-compose/cantabular-import/analysis
TR (top right)   : cd dp-recipe-api/import-recipes
BL (bottom left) : cd dp-compose/cantabular-import
BR (bottom right): run ‘docker stats’

Pre:
- Ensure docker resources are ok, especially memory allocated (can give 8G should be enough, may need little more as other containers are added)
- Update to latest version of dp-compose
- If not using mac, Update any usage of ‘usr/local/bin/bash’ to ‘/bin env bash’ or whatever in following (and other) scripts. There are at least 2 places to do this, and they are related to the final point:
- Setup Florence token:
	Setup the ‘get-florence-token.sh’ script (changing it’s ‘usr/local/bin/bash’ as above). Note this needs the password for the ‘florence@magicroundabout.ons.gov.uk’ user (so don’t forget it)
	Setup the ‘florence-token’ script  in ‘./helpers’ - this is a copy of the above script and needs the same changes

Now in the BL window, run the ‘run-cantabular-without-sudo.sh’ script. This might prompt for other repos to pull, especially when running for the first time.
Sometimes this will not work - do a Ctrl-C  and try again, but remember to do a ‘docker compose down’ first. Sometimes it may even necessitate a docker desktop re-start.
When everything is running successfully, then run ‘docker stats’ in the BR
Wait for everything to settle down, i.e. all containers to be running and the CPU usage dropping very low AND Florence CPU is below 5%

In TR window … run:

./import-recipes.sh mongodb://localhost:27017

(
This may fail to run, informing you that not all of the containers needed are running (typically 1 or more of the Kafka containers). If this happens, stop the stack - observing that all containers for the compose stack are stopped in the BR docker stats window and restart all the containers in the BL window.
)

When this is done, then in TL (analysis’), run ‘test-and-logs-by-id.sh X’, where X is the number of times to run the cantabular import.
This also copies all the logs out of the containers for analysis. Read the README and look at the various helper scripts in this folder.

Notes:
‘docker compose down’ will get rid of the logs; can use ‘portainer’ to get rid of unused volumes (portainer.io, available as a docker download).

If you pull a later code base of any of the apps that are being used, sometimes it does not seem to get rebuilt … if this happens, do the following:
Stop the compose stack
Docker compose down
Ensure all containers are stopped (see the docker stats window)
You may need to restart docker if not all containers will stop.
Delete all the unused containers related to the compose stack, using portioner &/or docker desktop.
Delete all volumes related to compose stack.
Delete the image(s) that won't rebuild properly (you will see that their build timestamp is not within the last minute or so).
Run: “docker builder prune --all” … as sometimes for zebedee the latest build date shows that it has not rebuilt and the ‘prune’ seems to finally get rid of image info that causes it to not build the actual latest.
