package main

import (
	"bufio"
	"fmt"
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

		var linesFound []string

		maxFirstFieldLength := 0

		// search through all container logs
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

		for _, line := range linesFound {
			fields := strings.Fields(line)
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
			printAndSave(idResultFile, fmt.Sprintf("     Job start time: %s", jobStart))

			printAndSave(idResultFile, fmt.Sprintf("   first event time: %s", firstTime))
			printAndSave(idResultFile, fmt.Sprintf("    last event time: %s", lastTime))

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

			printAndSave(idResultFile, fmt.Sprintf("Job execution time is: %d.%d seconds\n", diffNanoseconds/1000000000, diffNanoseconds%1000000000))
		}
		cerr := logFile.Close()
		if cerr != nil {
			fmt.Printf("problem closing: %s : %v\n", dockerLogName, cerr)
		}
		logFile, err = os.Open(dockerLogName)
		check(err)
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
