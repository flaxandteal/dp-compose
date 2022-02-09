package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	dockerLogName = "../tmp/all-container-logs.txt"
)

// this list MUST contain the names of the required services
var requiredServices = []string{
	"babbage",
	"dp-api-router",
	"dp-cantabular-api-ext",
	"dp-cantabular-csv-exporter",
	"dp-cantabular-dimension-api",
	"dp-cantabular-filter-flex-api",
	"dp-cantabular-metadata-exporter",
	"dp-cantabular-server",
	"dp-cantabular-xlsx-exporter",
	"dp-dataset-api",
	"dp-download-service",
	"dp-filter-api",
	"dp-frontend-dataset-controller",
	"dp-frontend-router",
	"dp-import-api",
	"dp-import-cantabular-dataset",
	"dp-import-cantabular-dimension-options",
	"dp-publishing-dataset-controller",
	"dp-recipe-api",
	"florence",
	"kafka-1",
	"kafka-2",
	"kafka-3",
	"minio",
	"mongodb",
	"postgres",
	"the-train",
	"vault",
	"zebedee",
	"zookeeper-1",
}

// It can take a while to read all container logs, so all logs are read into one file and then this file is
// scanned through repeatedly for required info by other apps.
func main() {
	fmt.Println("\nExtracting Docker Logs...")

	skip := false
	if len(os.Args) > 1 {
		param := os.Args[1]
		if reflect.TypeOf(param).Kind() == reflect.String {
			param = strings.ToLower(param)
			if param == "skip" {
				skip = true
			}
		}
	}

	if skip == false {
		count, serviceNames, err := getCantabularContainerCount()
		if err != nil {
			fmt.Printf("Error getting container count: %v\n", err)
			os.Exit(1)
		}
		maxContainersInJob := len(requiredServices)
		if count != maxContainersInJob {
			fmt.Printf("Incorrect number of Cantabular containers found.\nWanted: %d, found: %d\n... have you started the containers ?\n", maxContainersInJob, count)
			if count > 0 {
				listMissingServices(serviceNames)
			}
			os.Exit(2)
		}
		fmt.Printf("All %d containers present\n\n", maxContainersInJob)
	} else {
		fmt.Printf("    ****  Skipping checking if all containers present  ****\n\n")
	}

	err := createLogFileForAllDockerContainers()
	if err != nil {
		fmt.Printf("createLogFileForAllDockerContainers failed: %v\n", err)
		os.Exit(3)
	}
}

func getCantabularContainerCount() (int, []string, error) {
	var serviceNames []string

	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return 0, serviceNames, err
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return 0, serviceNames, err
	}

	cantabularContainersCount := 0

	for _, container := range containers {

		if strings.Contains(container.Names[0], "/cantabular-import-journey") {
			cantabularContainersCount++
			serviceNames = append(serviceNames, container.Labels["com.docker.compose.service"])
		}
	}

	return cantabularContainersCount, serviceNames, nil
}

func listMissingServices(serviceNames []string) {
	var serviceList []string = requiredServices
	//copy(serviceList, requiredServices)
	for _, foundService := range serviceNames {
		for index, service := range serviceList {
			if service == foundService {
				serviceList[index] = "" // remove found service
				break
			}
		}
	}

	// display the remaining service names that have not been found
	for _, service := range serviceList {
		if service != "" {
			fmt.Printf("Docker container not found: %s\n", service)
		}
	}
}

func createLogFileForAllDockerContainers() error {
	fmt.Printf("Getting Container Logs\n")
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	f, err := os.Create(dockerLogName)
	if err != nil {
		return err
	}
	defer func() {
		cerr := f.Close()
		if cerr != nil {
			fmt.Printf("problem closing: %s : %v\n", dockerLogName, cerr)
		}
	}()

	containersReadCount := 0

	for _, container := range containers {

		options := types.ContainerLogsOptions{ShowStdout: true, Timestamps: true}
		out, err := cli.ContainerLogs(ctx, container.ID, options)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(out)

		for scanner.Scan() {
			line := scanner.Text()

			if len(line) > 8 {
				// skip first 8 bytes as these are Docker's info about the line and are not printable characters
				strLine := string(line[8:])

				// remove colour highlighting codes, '\x1B' being the Escape code
				strLine = strings.ReplaceAll(strLine, "\x1B[34;1m", "")
				strLine = strings.ReplaceAll(strLine, "\x1B[32;1m", "")
				strLine = strings.ReplaceAll(strLine, "\x1B[36;1m", "")
				strLine = strings.ReplaceAll(strLine, "\x1B[33;1m", "")
				strLine = strings.ReplaceAll(strLine, "\x1B[30;1m", "")
				strLine = strings.ReplaceAll(strLine, "\x1B[0m", "")

				// prefix container's name
				strLine = container.Names[0] + " " + strLine + "\n"
				_, err := f.WriteString(strLine)
				if err != nil {
					fmt.Printf("error writing processed container line: %v\n", strLine)
					break
				}
			}
		}
		containersReadCount++
	}
	err = f.Sync()
	if err != nil {
		fmt.Printf("problem sync'ing writes out to: %s\n", dockerLogName)
	}

	if containersReadCount > 0 {
		fmt.Printf("Logs read OK\n")
	} else {
		return errors.New("no containers found to read ... have you started the containers")
	}
	return nil
}
