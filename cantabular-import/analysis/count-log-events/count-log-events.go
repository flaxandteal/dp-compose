package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	dockerLogName        = "../tmp/all-container-logs.txt"
	tmpFileName          = "../tmp/id.txt"
	plotDataFileName     = "../tmp/plot.txt"
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

	plotDataFile, err := os.OpenFile(plotDataFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
	check(err)
	defer func() {
		cerr := plotDataFile.Close()
		if cerr != nil {
			fmt.Printf("problem closing: %s : %v\n", plotDataFileName, cerr)
		}
	}()

	jobCount := 0
	idsFound := 0

	idScan := bufio.NewScanner(idTextFile)

	firstTime := "3333" // a year way in the future
	lastTime := "1111"  // a year way in the past

	var idList []string // to save id's for more comparisons later

	// for each job id
	for idScan.Scan() {
		fields := strings.Fields(idScan.Text())

		if len(fields) != 2 {
			fmt.Printf("error in id.txt file. expected time and id in line, but got %v\n", fields)
			os.Exit(1)
		}
		jobStart := fields[0]
		id := fields[1]

		idList = append(idList, id)

		// include the job start time in the determining of the very first time to ensure
		// any relevant events are considered.
		if jobStart < firstTime {
			firstTime = jobStart
		}

		jobCount++

		printAndSave(resultFile, fmt.Sprintf("Looking for id: %s\n", id))

		// search through all container logs
		_, err = logFile.Seek(0, io.SeekStart)
		check(err)
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
	}
	err = idScan.Err()
	check(err)

	// we now need to subtract a few milliseconds from the firsTime value found so that we grab the whole of the first event that contains the first id found.
	// this means that we may grab a few extra events before the desired event.

	f, ferr := time.Parse(time.RFC3339, firstTime) // time format with nanoseconds
	if ferr != nil {
		fmt.Printf("error in firstTime: %v\n", ferr)
	} else {
		f = f.Add(-(time.Second / 1000) * 2)
		// (the specific format chosen is to be compatible with the ones in the docker logs, and thus makes
		//  comparison of time's easily possible)
		firstTime = f.Format("2006-01-02T15:04:05.000000000Z")
	}

	// and similarly for the lastTime
	l, lerr := time.Parse(time.RFC3339, lastTime) // time format with nanoseconds
	if lerr != nil {
		fmt.Printf("error in firstTime: %v\n", lerr)
	} else {
		l = l.Add((time.Second / 1000) * 2)
		// (the specific format chosen is to be compatible with the ones in the docker logs, and thus makes
		//  comparison of time's easily possible)
		lastTime = l.Format("2006-01-02T15:04:05.000000000Z")
	}

	var linesFound []string // to store start and end of events lines between the time range plus lines that contain id's

	// Now look for events between firstTime and lastTime in all log files.
	// This is done in a simple mined manned taking into account that the format of the log lines
	// helps us identify the start of an event with just one opening curly bracket in the info field[2] of for example:
	/*
	 /cantabular-import-journey_dp-recipe-api_1 2021-07-30T09:33:52.503939175Z {
	*/
	if idsFound > 0 {
		eventsCountStarts := 0
		eventsCountEnds := 0
		wrapperCounts := 0
		eventsWithIds := 0

		maxFirstFieldLength := 0

		var validLines []string
		var validCount int

		// search through all container logs to retrieve event lines within time range
		_, err := logFile.Seek(0, io.SeekStart)
		check(err)

		scanner := bufio.NewScanner(logFile)
		for scanner.Scan() {
			strLine := string(scanner.Text())

			fields := strings.Fields(strLine)
			if len(fields) > 2 {
				// we have three fields
				if fields[1] >= firstTime && fields[1] <= lastTime {
					// the fields are within the time range
					validLines = append(validLines, strLine)
					validCount++
				}
			}
		}
		err = scanner.Err()
		check(err)

		for j := 0; j < validCount; j++ {
			strLine := validLines[j]
			fields := strings.Fields(strLine)

			if len(fields[0])+len(fields[1])+len(fields[2])+2 == len(strLine) {
				// the line has no leading spaces in field[2], that got stripped out by strings.Fields()
				if fields[2] == "{" {
					// and its only the opening curly bracket that we are interested in to indicate the start of a well formed event
					eventsCountStarts++
					if len(fields[0]) > maxFirstFieldLength {
						// save field width for later
						maxFirstFieldLength = len(fields[0])
					}

					var eventStr = "{"
					var idLine string
					// extract all lines of an event into one line
					j++
					for j < validCount {
						eLine := validLines[j]
						f := strings.Fields(eLine)
						if len(f) > 2 {
							// we have three fields
							if len(f[0])+len(f[1])+len(f[2])+2 == len(eLine) {
								// the line has no leading spaces in field[2], that got stripped out by strings.Fields()
								if f[2] == "}" {
									eventStr += "}"

									// a space on the end of the strings is the field delimeter
									offset := "0.0 " // an offset for Y axis that is prefixed to line indicating event is a wrapper and which it is
									if strings.Contains(eventStr, "http request received") {
										wrapperCounts++
										offset = "0.3 "
									}
									if strings.Contains(eventStr, "http request completed") {
										wrapperCounts++
										offset = "-0.3 "
									}

									kafkaType := "k=n "
									if strings.Contains(eventStr, "event received") {
										// got a consume message
										kafkaType = "k=c "
									}
									if strings.Contains(eventStr, "producing new cantabular dataset instance started event") {
										// got a produce message
										kafkaType = "k=p "
									}
									if strings.Contains(eventStr, "Triggering dimension options import") {
										// got a produce message that is about to produce 'multiple' messages
										kafkaType = "k=mp "
									}

									// examine event and do appropriate log: just line showing id or opening curly bracket
									if idLine != "" {
										linesFound = append(linesFound, kafkaType+offset+idLine)
										// An id may appear in an event more than once, so this counter
										// adds up just the events that have one or more id's in them
										eventsWithIds++
									} else {
										linesFound = append(linesFound, kafkaType+offset+strLine)
									}
									break
								}
							} else {
								for _, id := range idList {
									if strings.Contains(eLine, id) {
										idLine = eLine
										break
									}
								}

								eLine = strings.ReplaceAll(eLine, f[0], "")
								eLine = strings.ReplaceAll(eLine, f[1], "")

								for {
									if string(eLine[0]) != " " {
										break
									}
									eLine = strings.TrimPrefix(eLine, " ")
								}

								eventStr += eLine
							}
						}
						j++
					}
					j--
				}
				if fields[2] == "}" {
					// or the closing bracket
					eventsCountEnds++
				}
			}
		}

		// we now take the max of the two counts, just in case the time range that we checked between
		// has not captured one of the curly brackets
		maxEvents := eventsCountStarts
		if eventsCountEnds > maxEvents {
			maxEvents = eventsCountEnds
		}

		sort.SliceStable(linesFound, func(i, j int) bool {
			// extract and sort by the timestamp for each line
			fieldsi := strings.Fields(linesFound[i])
			fieldsj := strings.Fields(linesFound[j])
			return fieldsi[3] < fieldsj[3]
		})
		gotFirstTime := false
		maxLastTime := ""
		var maxDiff time.Duration
		maxDiffTime := ""

		var diffsFound []time.Duration // to store time differences between id's being searched for
		var serviceNames []string
		var eventHasId []bool
		var wrappedOffset []string
		var kafkaIndication []string
		var firstServiceName string
		var total time.Duration

		for _, line := range linesFound {
			fields := strings.Fields(line)

			if gotFirstTime {
				f, ferr := time.Parse(time.RFC3339, maxLastTime) // time format with nanoseconds
				if ferr != nil {
					fmt.Printf("error in maxLastTime: %v\n", ferr)
				}

				l, lerr := time.Parse(time.RFC3339, fields[3])
				if lerr != nil {
					fmt.Printf("error in fields[3]: %v\n", lerr)
				}

				if ferr == nil && lerr == nil {
					fTime := f.Local()
					lTime := l.Local()

					diffNanoseconds := lTime.Sub(fTime)
					total += diffNanoseconds
					diffsFound = append(diffsFound, diffNanoseconds)
					if diffNanoseconds > maxDiff {
						maxDiff = diffNanoseconds
						maxDiffTime = fields[3]
					}
					maxLastTime = fields[3]
					if fields[4] == "{" {
						eventHasId = append(eventHasId, false)
					} else {
						eventHasId = append(eventHasId, true)
					}
					printAndSave(resultFile, fmt.Sprintf("current: %.9f seconds", total.Seconds()))
					printAndSave(resultFile, fmt.Sprintf("diff: %.9f seconds", diffNanoseconds.Seconds()))

					serviceNames = append(serviceNames, fields[2])
					wrappedOffset = append(wrappedOffset, fields[1])
					kafkaIndication = append(kafkaIndication, fields[0])
				}
			} else {
				gotFirstTime = true
				maxLastTime = fields[3]
				firstServiceName = fields[2]
			}

			f1 := fields[2]
			for len(f1) < maxFirstFieldLength {
				// pad out the first field so that all timestamps are aligned
				f1 += " "
			}
			f1 += " "
			f1 += fields[3]
			f1 += line[len(fields[1])+1+len(fields[2])+1+len(fields[3]):]
			f1 = fields[1] + " " + f1 // prefix the wrapped offset value
			f1 = fields[0] + " " + f1 // prefix the kafka consume indication
			printAndSave(resultFile, fmt.Sprintf("%s", f1))
		}

		if len(serviceNames) > 0 {
			// save data for plotting
			// The first service name is effectively time '0'
			printAndSave(plotDataFile, fmt.Sprintf("k=n 0.0 %s 0.0000 true", firstServiceName))
			var timeOffest time.Duration
			for i := 0; i < len(diffsFound); i++ {
				timeOffest += diffsFound[i]
				printAndSave(plotDataFile, fmt.Sprintf("%s %s %s %.4f %v", kafkaIndication[i], wrappedOffset[i], serviceNames[i], timeOffest.Seconds(), eventHasId[i]))
			}
		}

		printAndSave(resultFile, fmt.Sprintf("max id execution time is: %.9f seconds, finishing at: %s\n", maxDiff.Seconds(), maxDiffTime))
		sort.SliceStable(diffsFound, func(i, j int) bool {
			// compare and sort by the durations
			return diffsFound[i] < diffsFound[j]
		})
		printAndSave(resultFile, fmt.Sprintf("diffs: %v\n", diffsFound))
		printAndSave(resultFile, fmt.Sprintf("len of diffs: %v\n", len(diffsFound)))

		// deduce how much of the overall time is taken by the ~10% of the largest diffs
		topTenStart := len(diffsFound) - (len(diffsFound)/10 + 1)
		var topTenTotal time.Duration
		for i := 0; i < len(diffsFound); i++ {
			if i >= topTenStart {
				topTenTotal += diffsFound[i]
			}
		}
		printAndSave(resultFile, fmt.Sprintf("Largest ~10%% of diffsFound adds up to: %v\n", topTenTotal))
		printAndSave(resultFile, fmt.Sprintf("Which is %v%% of the total\n", (100*topTenTotal.Nanoseconds())/total.Nanoseconds()))

		printAndSave(resultFile, fmt.Sprintf("Number of ID's found is: %d", idsFound))
		printAndSave(resultFile, fmt.Sprintf("Number of events with ID(s) is: %d", eventsWithIds))

		if firstTime != "" && lastTime != "" {
			printAndSave(resultFile, fmt.Sprintf(" first event time: %s", firstTime))
			printAndSave(resultFile, fmt.Sprintf("last event time: %s", lastTime))

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

			printAndSave(resultFile, fmt.Sprintf("Job(s) execution time is: %.9f seconds\n", diffNanoseconds.Seconds()))
		}

		printAndSave(resultFile, fmt.Sprintf("Total events found (within first and last times): %d\n", maxEvents))

		printAndSave(resultFile, fmt.Sprintf("wrapperCounts: %d", wrapperCounts))

		printAndSave(resultFile, fmt.Sprintf("Total Jobs: %d", jobCount))
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
