package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

const (
	plotDataFileName = "../tmp/plot.txt"
	idOffsetY        = 0.3 // Y axis offset of the ID event to separate overlaping lines for clarity
)

type plotXY struct {
	x    float64
	y    float64
	isId bool
}

// read extracted events and do a simple plot
func main() {
	plotDataFile, err := os.Open(plotDataFileName)
	check(err)
	defer func() {
		cerr := plotDataFile.Close()
		if cerr != nil {
			fmt.Printf("problem closing: %s : %v\n", plotDataFileName, cerr)
		}
	}()

	// first read through file to get container names and thus the number of containers
	// to give each event a position on the Y axis
	plotDataScan := bufio.NewScanner(plotDataFile)

	whatContainers := make(map[string]int)

	var containerIndex = 1 // start at 1 above Y axis

	// for each event line extract container name and buld up a map of individual container names
	for plotDataScan.Scan() {
		fields := strings.Fields(plotDataScan.Text())

		if len(fields) != 4 {
			fmt.Printf("error in plot.txt file. expected 'Y offset', 'service name', 'relative time' and 'is ID flag' in line, but got %v\n", fields)
			os.Exit(1)
		}
		containerName := fields[1]

		if _, ok := whatContainers[containerName]; !ok {
			fmt.Printf("Name: %s, index: %d\n", containerName, containerIndex)
			whatContainers[containerName] = containerIndex
			containerIndex++
		}
	}
	err = plotDataScan.Err()
	check(err)

	var plotData []plotXY

	// extract event times into X 'time' and Y 'container name'
	_, err = plotDataFile.Seek(0, io.SeekStart)
	check(err)

	scanner := bufio.NewScanner(plotDataFile)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())

		yOffset := fields[0]
		containerName := fields[1]
		offset := fields[2]
		hasId := fields[3]

		var pData plotXY

		pData.x, err = strconv.ParseFloat(offset, 64)
		check(err)
		yOffseFloat, err := strconv.ParseFloat(yOffset, 64)
		check(err)
		pData.y = float64(whatContainers[containerName]) + yOffseFloat
		pData.isId, err = strconv.ParseBool(hasId)
		check(err)

		plotData = append(plotData, pData)
	}
	err = scanner.Err()
	check(err)

	diffsPlot, totalEvents, nofIds, err := plotAll(plotData)
	check(err)

	diffsPlot.Title.Text = fmt.Sprintf("Events for containers - timeline, spanning: %d events - of which: %d have import job Ids", totalEvents, nofIds)
	diffsPlot.X.Label.Text = "time in seconds"
	diffsPlot.Y.Label.Text = "service / container name"

	cNames := make([]string, len(whatContainers)+1)

	for k, v := range whatContainers {
		k = strings.ReplaceAll(k, "/cantabular-import-journey_", "") // this might best be in code that created files
		k = strings.ReplaceAll(k, "_1", "")                          // this might best be in code that created files
		cNames[v] = k
	}

	diffsPlot.NominalY(cNames...)

	err = diffsPlot.Save(100*vg.Centimeter, 20*vg.Centimeter, "diffsPlot.svg")
	check(err)
}

func plotAll(plotData []plotXY) (*plot.Plot, int, int, error) {
	var nofIds int
	p := plot.New()

	// create 1st plot of all events
	points := make(plotter.XYs, len(plotData))
	for i, val := range plotData {
		points[i].X = val.x
		points[i].Y = val.y
		if val.isId {
			nofIds++
		}
	}

	// create 2nd plot line of just the events with the job ID to overlay on all events
	idPoints := make(plotter.XYs, nofIds)
	var index int
	for _, val := range plotData {
		if val.isId {
			idPoints[index].X = val.x
			idPoints[index].Y = val.y + idOffsetY
			index++
		}
	}

	if err := plotutil.AddLinePoints(p, "All events:", points, "Events with Ids:", idPoints); err != nil {
		return nil, 0, 0, err
	}
	return p, len(plotData), nofIds, nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
