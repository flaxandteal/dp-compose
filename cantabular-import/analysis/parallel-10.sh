#!/usr/bin/env bash

# kick off ten jobs in parallel
for i in {1..10}; do
    (
        echo "$i"
        ./start-analysis.sh
    ) &
done

# no more jobs to be started but wait for pending jobs
# (all need to be finished)
wait

echo "all done"
