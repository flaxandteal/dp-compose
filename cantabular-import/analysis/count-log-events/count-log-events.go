package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	dockerLogName             = "../tmp/all-container-logs.txt"
	tmpFileName               = "../tmp/id.txt"
	plotDataFileName          = "../tmp/plot.txt"
	countResultsFileName      = "count-log-events-results.txt"
	instanceEventsFileName    = "instance-events.txt"
	filterHealthEtcFromEvents = true
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

	instanceEventsFile, err := os.OpenFile(instanceEventsFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
	check(err)
	defer func() {
		cerr := instanceEventsFile.Close()
		if cerr != nil {
			fmt.Printf("problem closing: %s : %v\n", instanceEventsFileName, cerr)
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
		fmt.Printf("error in lastTime: %v\n", lerr)
	} else {
		l = l.Add((time.Second / 1000) * 500)
		// (the specific format chosen is to be compatible with the ones in the docker logs, and thus makes
		//  comparison of time's easily possible)
		lastTime = l.Format("2006-01-02T15:04:05.000000000Z")
	}

	var linesFound []string        // to store start and end of events lines between the time range plus lines that contain id's
	var allInstanceEvents []string // to save off a chronological list of all event information pertaining to the instance

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

									kafkaType := "k=n " // initialise to indicate 'not' a kafka log event
									if strings.Contains(eventStr, "event received") ||
										strings.Contains(eventStr, "handling pre-publish") { // dp-dataset-exporter-xlsx : handleFullDownloadMessage()
										// this is a consume message
										kafkaType = "k=c "
									}
									if strings.Contains(eventStr, "producing new cantabular dataset instance started event") ||
										strings.Contains(eventStr, "all dimensions in instance have been completely processed and kafka message has been sent") ||
										strings.Contains(eventStr, "producing common output created event") {
										// this is a produce message
										kafkaType = "k=p "
									}
									if strings.Contains(eventStr, "Triggering dimension options import") {
										// this is a produce message that is about to produce 'multiple' messages
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

					serviceName := strings.ReplaceAll(fields[0], "/cantabular-import-journey_", "") // this might best be in code that created files
					serviceName = strings.ReplaceAll(serviceName, "_1", "")                         // this might best be in code that created files
					eventLine := fmt.Sprintf("%40s: %s\n", serviceName, eventStr)
					allInstanceEvents = append(allInstanceEvents, eventLine)
				}
				if fields[2] == "}" {
					// or the closing bracket
					eventsCountEnds++
				}
			} else {
				// handle non standard Sensible code log lines
				if len(fields) > 4 {
					strLine = fields[0] + " " + fields[1]
					strLine += " {" // ditch the rest of sensible code / cantabular log line and replace with '{' to indicate no job id in line
					linesFound = append(linesFound, "k-n 0.0 "+strLine)
				}
			}
		}

		sort.SliceStable(allInstanceEvents, func(i, j int) bool {
			// extract and sort by the timestamp for each line
			fieldsi := strings.Fields(allInstanceEvents[i])
			fieldsj := strings.Fields(allInstanceEvents[j])
			var fi, fj []string
			// zebedee logs split differently to all other services, so require a different check:
			if strings.Contains(allInstanceEvents[i], "zebedee: {") {
				fi = strings.Split(fieldsi[3], ",")
			} else {
				fi = strings.Split(fieldsi[2], ",")

			}
			if strings.Contains(allInstanceEvents[j], "zebedee: {") {
				fj = strings.Split(fieldsj[3], ",")
			} else {
				fj = strings.Split(fieldsj[2], ",")
			}
			return fi[0] < fj[0]
		})

		// create regex to remove timestamp string: "created_at": "2021-12-01T09:33:58.1004706Z",
		reg := regexp.MustCompile(`"created_at": "[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1])T(2[0-3]|[01][0-9]):[0-5][0-9].*Z",`)

		// save event lines
		for _, s := range allInstanceEvents {
			// for zebedee log lines, remove space before colon so that all the created at timestamps line up vertically in file
			e := strings.ReplaceAll(s, "zebedee: {\"created_at\" :", "zebedee: {\"created_at\":")
			if filterHealthEtcFromEvents {
				// filter out events that don't have any useful info to following the import/export process
				if strings.Contains(e, `","instanceID": "","location": "","payload": null,"type": ""}},"event": "client log","namespace": "florence","severity": 3}`) ||
					strings.Contains(e, `","event": "executing identity check middleware","namespace":`) ||
					strings.Contains(e, `","data": {"auth_token": {},"florence_token": {},"is_service_request": true,"is_user_request": true,"url": "http://zebedee:8082/identity"},"event": "calling AuthAPI to authenticate caller identity","namespace":`) ||
					strings.Contains(e, `","http" : {"method" : "GET","path" : "/identity","scheme" : "http","host" : "zebedee","port" : 8082}}`) ||
					strings.Contains(e, `","http" : {"method" : "GET","path" : "/health","scheme" : "http","host" : "zebedee","port" : 8082}}`) ||
					//strings.Contains(e, `","namespace" : "zebedee-reader","severity" : 3,"event" : "request received","trace_id" `) ||
					strings.Contains(e, `","event": "identity client check request completed successfully invoking downstream http handler","namespace":`) ||
					strings.Contains(e, `","data": {"action": "retrieving service health","method": "GET","uri": "http://zebedee:8082"},"event": "Making request to service: Zebedee","namespace":`) ||
					strings.Contains(e, `"data": {"action": "retrieving service health","method": "GET","uri": "http:`) ||
					strings.Contains(e, `","event": "http request received","http": {"method": "GET","path": "/health","started_at": "`) ||
					strings.Contains(e, `","method": "GET","path": "/health","response_content_length":`) ||
					strings.Contains(e, `","event": "florence access token header not found attempting to find access token cookie","namespace": "`) ||
					strings.Contains(e, `","event": "florence access token cookie not found in request","namespace": "`) ||
					strings.Contains(e, `","data": {"auth_token": {},"is_service_request": true,"is_user_request": false,"url": "http://zebedee:8082/identity"},"event": "calling AuthAPI to authenticate caller identity","namespace":`) ||
					strings.Contains(e, `","data": {"auth_token": {},"caller_identity": "","is_service_request": true,"is_user_request": false,"url": "http://zebedee:8082/identity","user_identity": ""},"event": "caller identity retrieved setting context values","namespace":`) ||
					strings.Contains(e, `","data": {"service": "image-api"},"errors": [{"message": "Get \"http://localhost:24700/health\": dial tcp 127.0.0.1:24700: connect: connection refused","stack_trace":`) ||
					strings.Contains(e, `","event": "http request received","http": {"method": "GET","path": "/v1/health","started_at":`) ||
					strings.Contains(e, `","data": {"florence_token": {},"is_service_request": false,"is_user_request": true,"url": "http://zebedee:8082/identity"},"event": "calling AuthAPI to authenticate caller identity","namespace": "dp-dataset-api","severity": 3}`) ||
					strings.Contains(e, `","method": "GET","path": "/v1/health","response_content_length":`) ||
					strings.Contains(e, `,"event": "getDataset endpoint: caller not authorised returning dataset","namespace":`) ||
					strings.Contains(e, `","data": {"service": "filter-api"},"errors": [{"message": "Get \"http:`) ||
					strings.Contains(e, `"event": "service auth token request header is not found","namespace":`) ||
					// the following don't aid studying successful event sequences, so they are also filtered out
					strings.Contains(e, `"status_code": 0},"namespace": "dp-api-router","severity": 3,"trace_id":`) ||
					//strings.Contains(e, `"status_code": 0},"namespace": "dp-dataset-api","severity": 3`) ||
					strings.Contains(e, `"status_code": 0},"namespace": "florence"`) ||
					strings.Contains(e, `"status_code": 0},"namespace": "dp-import-api","severity": 3`) ||
					strings.Contains(e, `"status_code": 0},"namespace": "recipe-api","severity": 3`) ||
					strings.Contains(e, `"status_code": 0},"namespace": "dp-publishing-dataset-controller","severity": 3`) ||
					strings.Contains(e, `"status_code": 200},"namespace": "`) {
					continue
				}
			}

			// As the list is in correct timestamp order, by removing the timestamp ... the application 'meld' is able to matchup and
			// nicely visually compare the resulting file with previously saved/renamed files (without removing the timestamp there are too may
			// differences for meld to funciton as desired)
			res := reg.ReplaceAllString(e, "")
			_, err := fmt.Fprintf(instanceEventsFile, res)
			check(err)
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
