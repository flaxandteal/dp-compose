### Helper scripts ###

Find here some scripts to help with working with the Cantabular import process

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
