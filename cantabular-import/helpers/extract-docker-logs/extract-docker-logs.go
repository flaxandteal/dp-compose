package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	dockerLogName = "../tmp/all-container-logs.txt"
)

// It can take a while to read all container logs, so all logs are read into one file and then this file is
// scanned through repeatedly for required info by other apps.
func main() {
	err := createLogFileForAllDockerContainers()
	if err != nil {
		fmt.Printf("createLogFileForAllDockerContainers failed: %v\n", err)
		os.Exit(1)
	}
}

func createLogFileForAllDockerContainers() error {
	fmt.Printf("Getting all Container Logs\n")
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
		return errors.New("no containers found to read ... have you started the containers ?")
	}
	return nil
}
