#!/usr/bin/env bash

set -e # 'e' to stop on error (non zero return from script)

N=1
SKIP=NO

if [[ -n $1 ]]; then
    N=$1
fi

if [[ -n $2 ]]; then
    SKIP=$2
fi

for i in $(seq 1 "$N");do
    printf "===================================================================\n\n"
    printf "Running integration test number: %s  of %s\n\n" "$i" "$N"

    [[ -f tmp/id.txt ]] && rm tmp/id.txt

    cd full-import-export
    result=0
    go run -race main.go "$SKIP" || result=$?
    cd -

    if [[ $result -ne 0 ]]; then
        cd extract-docker-logs
        go run -race extract-docker-logs.go "$SKIP" || true
        cd -
        printf "===============\n"
        printf "Test failed - So analysis skipped but just extracted all container logs for examination\n"

        exit $result
    fi

    cd extract-docker-logs
    go run -race extract-docker-logs.go
    cd -

    cd count-log-events
    go run -race count-log-events.go
    cd -

    python3 report-errors.py
done

echo "Completed: $N integration tests OK"
