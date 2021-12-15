#!/usr/local/bin/bash
rm tmp/id.txt

set -ex # 'e' to stop on error (non zero return from script), 'x' to show command as it runs

cd full-import-export
go run main.go
cd ..

cd extract-docker-logs
go run extract-docker-logs.go
cd ..

cd count-log-events
go run count-log-events.go
cd ..
