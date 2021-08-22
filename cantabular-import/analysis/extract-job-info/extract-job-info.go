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
	dockerLogName          = "../tmp/all-container-logs.txt"
	idFileName             = "../tmp/id.txt"
	extractResultsFileName = "extract-job-info-results.txt"
)

// Scan through previously extracted log file repeatedly looking for and getting info
// for each job id. This saves a lot of time comparred to re-reading all of the docker logs
// for each id.
func main() {
	idTextFile, err := os.Open(idFileName)
	check(err)
	defer func() {
		cerr := idTextFile.Close()
		if cerr != nil {
			fmt.Printf("problem closing: %s : %v\n", idFileName, cerr)
		}
	}()

	idResultFile, err := os.OpenFile(extractResultsFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
	check(err)
	defer func() {
		cerr := idResultFile.Close()
		if cerr != nil {
			fmt.Printf("problem closing: %s : %v\n", extractResultsFileName, cerr)
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

	idScan := bufio.NewScanner(idTextFile)

	// for each job id
	for idScan.Scan() {
		fields := strings.Fields(idScan.Text())

		if len(fields) != 2 {
			fmt.Printf("error in id.txt file. expected time and id in line, but got %v\n", fields)
			break
		}
		jobStart := fields[0]
		id := fields[1]

		jobCount++

		printAndSave(idResultFile, fmt.Sprintf("Looking for id: %s\n", id))

		firstTime := "3333" // a year way in the future
		lastTime := "1111"  // a year way in the past
		idsFound := 0

		var linesFound []string        // to store lines that contain the 'id' being searched for
		var diffsFound []time.Duration // to store time differences between id's being searched for

		maxFirstFieldLength := 0

		// search through all container logs
		_, err := logFile.Seek(0, io.SeekStart)
		check(err)

		scanner := bufio.NewScanner(logFile)
		for scanner.Scan() {
			strLine := string(scanner.Text())

			// for the specific 'id'
			if strings.Contains(strLine, id) {
				idsFound++
				fields := strings.Fields(strLine)
				if len(fields) > 1 {
					// save line to sort all lines for an id by time
					linesFound = append(linesFound, strLine)
					if len(fields[0]) > maxFirstFieldLength {
						// save field width for later
						maxFirstFieldLength = len(fields[0])
					}
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

		sort.SliceStable(linesFound, func(i, j int) bool {
			// extract the timestamp for each line
			fieldsi := strings.Fields(linesFound[i])
			fieldsj := strings.Fields(linesFound[j])
			return fieldsi[1] < fieldsj[1]
		})

		gotFirstTime := false
		maxLastTime := ""
		var maxDiff time.Duration
		maxDiffTime := ""

		for _, line := range linesFound {
			fields := strings.Fields(line)

			if gotFirstTime {
				f, ferr := time.Parse(time.RFC3339, maxLastTime) // time format with nanoseconds
				if ferr != nil {
					fmt.Printf("error in maxLastTime: %v\n", ferr)
				}

				l, lerr := time.Parse(time.RFC3339, fields[1])
				if lerr != nil {
					fmt.Printf("error in fields[1]: %v\n", lerr)
				}

				if ferr == nil && lerr == nil {
					fTime := f.Local()
					lTime := l.Local()

					diffNanoseconds := lTime.Sub(fTime)
					diffsFound = append(diffsFound, diffNanoseconds)
					if diffNanoseconds > maxDiff {
						maxDiff = diffNanoseconds
						maxDiffTime = fields[1]
					}
					maxLastTime = fields[1]
					printAndSave(idResultFile, fmt.Sprintf("time since last id: %.9f seconds", diffNanoseconds.Seconds()))
				}
			} else {
				gotFirstTime = true
				maxLastTime = fields[1]
			}

			f1 := fields[0]
			for len(f1) < maxFirstFieldLength {
				// pad out the first field so that all timestamps are aligned
				f1 = f1 + " "
			}
			f1 += " "
			f1 += fields[1]
			f1 += line[len(fields[0])+1+len(fields[1]):]
			printAndSave(idResultFile, fmt.Sprintf("%s", f1))
		}

		printAndSave(idResultFile, fmt.Sprintf("Number of ID's found is: %d", idsFound))
		if firstTime != "" && lastTime != "" && idsFound > 0 {
			printAndSave(idResultFile, fmt.Sprintf(" Job start time: %s", jobStart))

			printAndSave(idResultFile, fmt.Sprintf(" first event time: %s", firstTime))
			printAndSave(idResultFile, fmt.Sprintf("last event time: %s", lastTime))
			if idsFound > 1 {
				printAndSave(idResultFile, fmt.Sprintf("max id execution time is: %.9f seconds, finishing at: %s\n", maxDiff.Seconds(), maxDiffTime))
				sort.SliceStable(diffsFound, func(i, j int) bool {
					// compare the durations
					return diffsFound[i] < diffsFound[j]
				})
				printAndSave(idResultFile, fmt.Sprintf("diffs: %v\n", diffsFound))
				printAndSave(idResultFile, fmt.Sprintf("len of diffs: %v\n", len(diffsFound)))

				// deduce how much of the overall time is taken by the ~10% of the largest diffs
				topTenStart := len(diffsFound) - (len(diffsFound)/10 + 1)
				var total time.Duration
				var topTenTotal time.Duration
				for i := 0; i < len(diffsFound); i++ {
					total += diffsFound[i]
					if i >= topTenStart {
						topTenTotal += diffsFound[i]
					}
				}
				printAndSave(idResultFile, fmt.Sprintf("Largest ~10%% of diffsFound adds up to: %v\n", topTenTotal))
				printAndSave(idResultFile, fmt.Sprintf("Which is %v%% of the total\n", (100*topTenTotal.Nanoseconds())/total.Nanoseconds()))
			}

			f, ferr := time.Parse(time.RFC3339, firstTime) // time format with nanoseconds
			if ferr != nil {
				fmt.Printf("error in firstTime: %v\n", ferr)
			}

			l, lerr := time.Parse(time.RFC3339, lastTime)
			if lerr != nil {
				fmt.Printf("error in lastTime: %v\n", lerr)
			}

			if ferr == nil && lerr == nil {
				fTime := f.Local()
				lTime := l.Local()

				diffNanoseconds := lTime.Sub(fTime)

				printAndSave(idResultFile, fmt.Sprintf("Job execution time is: %.9f seconds\n", diffNanoseconds.Seconds()))
			}
		}
	}

	printAndSave(idResultFile, fmt.Sprintf("Total Jobs: %d", jobCount))

	err = idScan.Err()
	check(err)
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
