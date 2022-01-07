#!/usr/local/bin/bash

set -e # 'e' to stop on error (non zero return from script)

N="1"

if [ -n "$1" ]
then
    N=$1
fi

for i in $(seq 1 "$N")
do
    printf "===================================================================\n\n"
    printf "Running integration test number: %s  of %s\n\n" "$i" "$N"

    rm tmp/id.txt

    cd full-import-export
    go run main.go
    cd ..

    cd extract-docker-logs
    go run extract-docker-logs.go
    cd ..

    cd count-log-events
    go run count-log-events.go
    cd ..

    python3 report-errors.py
done

echo "Completed: $N integration tests OK"
