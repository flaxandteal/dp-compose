package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

const (
	plotDataFileName   = "../tmp/plot.txt"
	plotOutputFileName = "diffsPlot.svg"
	idOffsetY          = 0.3 // Y axis offset of the ID event to separate overlaping lines for clarity
)

type plotXY struct {
	x float64

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

	var index int // just to have different value

	// for each event line extract container name and buld up a map of individual container names
	for plotDataScan.Scan() {
		fields := strings.Fields(plotDataScan.Text())

		if len(fields) != 4 {
			fmt.Printf("error in plot.txt file. expected 'Y offset', 'service name', 'relative time' and 'is ID flag' in line, but got %v\n", fields)
			os.Exit(1)
		}
		containerName := fields[1]

		if _, ok := whatContainers[containerName]; !ok {
			fmt.Printf("Name: %s, index: %d\n", containerName, index)
			whatContainers[containerName] = index
			index++
		}
	}
	err = plotDataScan.Err()
	check(err)

	// alphabetically sort container names
	cNames := make([]string, len(whatContainers))

	for k, value_index := range whatContainers {
		cNames[value_index] = k
	}

	sort.Strings(cNames)

	whatContainersSorted := make(map[string]int)

	yOffset := 1 // start at 1 above Y axis (to help with better display of any minor applied offsets for wrapping http request midleware events)
	// extract names in descending order
	for i := len(cNames) - 1; i >= 0; i-- {
		fmt.Printf("Name: %s, Y axis offset: %d\n", cNames[i], yOffset)
		// and assign Y axis to try to maintain consistency between different plots
		whatContainersSorted[cNames[i]] = yOffset
		yOffset++
	}

	type request struct {
		x        float64
		y        float64
		gotFirst bool
	}
	requestMarkers := make([]request, len(cNames)+yOffset)

	type segment struct {
		startX float64
		startY float64
		endX   float64
		endY   float64
	}
	var segmentList []segment

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
		pData.y = float64(whatContainersSorted[containerName]) + yOffseFloat
		pData.isId, err = strconv.ParseBool(hasId)
		check(err)

		if yOffseFloat < -0.01 || yOffseFloat > 0.01 {
			if requestMarkers[whatContainersSorted[containerName]].gotFirst == false && yOffseFloat > 0.01 {
				requestMarkers[whatContainersSorted[containerName]].x = pData.x
				requestMarkers[whatContainersSorted[containerName]].y = pData.y
				requestMarkers[whatContainersSorted[containerName]].gotFirst = true
			} else if requestMarkers[whatContainersSorted[containerName]].gotFirst == true && yOffseFloat < -0.01 {
				requestMarkers[whatContainersSorted[containerName]].gotFirst = false
				// build segment info and add to list
				var seg segment
				seg.startX = requestMarkers[whatContainersSorted[containerName]].x
				seg.startY = requestMarkers[whatContainersSorted[containerName]].y + (yOffseFloat / 2)
				seg.endX = pData.x
				seg.endY = pData.y - (yOffseFloat / 2)
				segmentList = append(segmentList, seg)
			}
		}

		plotData = append(plotData, pData)
	}
	err = scanner.Err()
	check(err)

	diffsPlot, totalEvents, nofIds, err := plotAll(plotData)
	check(err)

	for _, seg := range segmentList {
		points := make(plotter.XYs, 5)

		// build segment shape ... !!! change to a bounding box with y's 0.1 higher and lower
		points[0].X = seg.startX
		points[0].Y = seg.startY

		points[1].X = seg.startX // draw down left side of box
		points[1].Y = seg.endY

		points[2].X = seg.endX // draw bottom of box
		points[2].Y = seg.endY

		points[3].X = seg.endX // draw right side of box
		points[3].Y = seg.startY

		points[4].X = seg.startX // close top side of box
		points[4].Y = seg.startY

		// add shape in 'blue' colour
		l, err := plotter.NewLine(points)
		check(err)
		l.Color = plotutil.Color(2)
		l.Dashes = plotutil.Dashes(0)
		diffsPlot.Add(l)
	}

	diffsPlot.Title.Text = fmt.Sprintf("Events for containers - timeline, spanning: %d events - of which: %d have import job Ids\nBlue boxes show events wrapped by middle ware http request (received & completed)", totalEvents, nofIds)
	diffsPlot.X.Label.Text = "time in seconds"
	diffsPlot.Y.Label.Text = "service / container name"

	cNamesYaxis := make([]string, len(whatContainersSorted)+1)

	for k, value_index := range whatContainersSorted {
		k = strings.ReplaceAll(k, "/cantabular-import-journey_", "") // this might best be in code that created files
		k = strings.ReplaceAll(k, "_1", "")                          // this might best be in code that created files
		cNamesYaxis[value_index] = k
	}

	diffsPlot.NominalY(cNamesYaxis...)

	err = diffsPlot.Save(100*vg.Centimeter, 20*vg.Centimeter, plotOutputFileName)
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
