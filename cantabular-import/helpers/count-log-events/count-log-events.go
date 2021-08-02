package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	dockerLogName        = "../tmp/all-container-logs.txt"
	tmpFileName          = "../tmp/id.txt"
	countResultsFileName = "count-log-events-results.txt"
)

// Using pre-read docker logs file, for all job id's deduce the combined first and last times in the log files.
// Then rescan all log files between the first and last times and count the number of events that
// are spread over many lines in a pretty print style manner of json format.
// Counts are shown and saved.
// If you are using this app to look at event counts for multiple Job ID's then ensure you ran ALL the Job's
// back to back or in parallel and then extracted the log files.
// NOTE: To get a clean test run, you need to stop the docker containers with 'docker-compose down' and
// then run up the containers again and do the recipe import thing to set up mongo and then your
// start-import(s) and then extract the logs, then run this app to get analysis.
func main() {
	idTextFile, err := os.Open(tmpFileName)
	check(err)
	defer func() {
		cerr := idTextFile.Close()
		if cerr != nil {
			fmt.Printf("problem closing: %s : %v\n", tmpFileName, cerr)
		}
	}()

	resultFile, err := os.OpenFile(countResultsFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
	check(err)
	defer func() {
		cerr := resultFile.Close()
		if cerr != nil {
			fmt.Printf("problem closing: %s : %v\n", countResultsFileName, cerr)
		}
	}()

	logFile, err := os.Open(dockerLogName)
	check(err)
	defer func() {
		cerr := logFile.Close()
		if cerr != nil {
			fmt.Printf("problem closing: %s : %v\n", dockerLogName, cerr)
		}
	}()

	jobCount := 0
	idsFound := 0

	idScan := bufio.NewScanner(idTextFile)

	firstTime := "3333" // a year way in the future
	lastTime := "1111"  // a year way in the past

	// for each job id
	for idScan.Scan() {
		fields := strings.Fields(idScan.Text())

		if len(fields) != 2 {
			fmt.Printf("error in id.txt file. expected time and id in line, but got %v\n", fields)
			os.Exit(1)
		}
		jobStart := fields[0]
		id := fields[1]

		// include the job start time in the determining of the very first time to ensure
		// any relevant events are considered.
		if jobStart < firstTime {
			firstTime = jobStart
		}

		jobCount++

		printAndSave(resultFile, fmt.Sprintf("Looking for id: %s\n", id))

		// search through all container logs
		scanner := bufio.NewScanner(logFile)
		for scanner.Scan() {
			strLine := string(scanner.Text())

			if strings.Contains(strLine, id) {
				idsFound++
				fields := strings.Fields(strLine)
				if len(fields) > 1 {
					if fields[1] < firstTime {
						firstTime = fields[1]
					}
					if fields[1] > lastTime {
						lastTime = fields[1]
					}
				}
			}
		}
		err = scanner.Err()
		check(err)

		cerr := logFile.Close()
		if cerr != nil {
			fmt.Printf("problem closing: %s : %v\n", dockerLogName, cerr)
		}
		logFile, err = os.Open(dockerLogName)
		check(err)
	}

	err = idScan.Err()
	check(err)

	printAndSave(resultFile, fmt.Sprintf("Number of ID's found is: %d", idsFound))
	if firstTime != "" && lastTime != "" && idsFound > 0 {
		printAndSave(resultFile, fmt.Sprintf("   first event time: %s", firstTime))
		printAndSave(resultFile, fmt.Sprintf("    last event time: %s", lastTime))

		f, err := time.Parse(time.RFC3339, firstTime) // time format with nanoseconds
		if err != nil {
			fmt.Println(err)
		}
		fTime := f.Local()

		l, err := time.Parse(time.RFC3339, lastTime)
		if err != nil {
			fmt.Println(err)
		}
		lTime := l.Local()

		diffNanoseconds := lTime.Sub(fTime)

		printAndSave(resultFile, fmt.Sprintf("Job(s) execution time is: %d.%d seconds\n", diffNanoseconds/1000000000, diffNanoseconds%1000000000))
	}

	printAndSave(resultFile, fmt.Sprintf("Total Jobs: %d", jobCount))

	// Now look for events between firstTime and lastTime in all log files.
	// This is done in a simple mined manned taking into account that the format of the log lines
	// helps us identify the start of an event with just one opening curly bracket in the info field[2] of for example:
	/*
	   /cantabular-import-journey_dp-recipe-api_1 2021-07-30T09:33:52.503939175Z {
	*/
	if idsFound > 0 {
		logFile, err := os.Open(dockerLogName)
		check(err)
		defer func() {
			cerr := logFile.Close()
			if cerr != nil {
				fmt.Printf("problem closing: %s : %v\n", dockerLogName, cerr)
			}
		}()

		eventsCountStarts := 0
		eventsCountEnds := 0

		// search through all container logs
		scanner := bufio.NewScanner(logFile)
		for scanner.Scan() {
			strLine := string(scanner.Text())

			fields := strings.Fields(strLine)
			if len(fields) > 2 {
				// we have thre fields
				if fields[1] >= firstTime && fields[1] <= lastTime {
					// the fields are within the time range
					if len(fields[0])+len(fields[1])+len(fields[2])+2 == len(strLine) {
						// the line has no leading spaces in field[2], that got stripped out by strings.Fields()
						if fields[2] == "{" {
							// and its only the opening curly bracket that we are interested in
							eventsCountStarts++
						}
						if fields[2] == "}" {
							eventsCountEnds++
						}
					}
				}
			}
		}

		// we now take the max of the two counts, just in case the time range that we checked between
		// has not captured one of the curly brackets
		maxEvents := eventsCountStarts
		if eventsCountEnds > maxEvents {
			maxEvents = eventsCountEnds
		}

		printAndSave(resultFile, fmt.Sprintf("Total events found: %d\n", maxEvents))
	}
}

func printAndSave(file *os.File, line string) {
	_, err := fmt.Fprintf(file, "%s\n", line)
	check(err)
	fmt.Println(line)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
