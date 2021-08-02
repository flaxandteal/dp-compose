package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type PutJobRequest struct {
	Links Links  `json:"links"`
	State string `json:"state"`
}

type PostJobResponse struct {
	ID    string `json:"id"`
	Links Links  `json:"links"`
}

type Links struct {
	Instances []Link `json:"instances"`
	Self      Link   `json:"self"`
}

type Link struct {
	ID   string `json:"id"`
	HRef string `json:"href"`
}

var (
	importAPIHost = "http://localhost:21800"
	recipeID      = "38542eb6-d3a6-4cc4-ba3f-32b25f23223a"

	httpClient = &http.Client{}
)

const (
	tmpFileName = "../tmp/id.txt"
)

const (
	maxContainersInJob = 11 // adjust this to suite the number of continers docker-compose runs up
	maxRuns            = 2  // number of times to run up containers, perform integration test and stop containers
)

func main() {

	for testCount := 0; testCount < maxRuns; testCount++ {
		fmt.Printf("Runnning test: %d\n\n", testCount)

		startContainers()

		for i := 0; i < 60; i++ {
			count, err := getContainerCount()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("Seconds: %03d   Container count: %d\n", i, count)
			time.Sleep(1 * time.Second)
			if count == maxContainersInJob {
				break
			}
		}

		// we should now check health of apps, but it seemd OK to just do a delay for these tests
		fmt.Println("Pasuing 15 seconds")
		time.Sleep(15 * time.Second)

		fmt.Printf("Doing Import ...\n")
		for attempts := 0; attempts <= 5; attempts++ {
			err := doImport()
			if err != nil {
				fmt.Println(err)
				if attempts < 5 {
					fmt.Println("Pasuing 5 seconds, and trying import again")
					time.Sleep(5 * time.Second)
				} else {
					fmt.Printf("Import failing ...\n")
					fmt.Println("Pasuing 5 seconds")
					time.Sleep(5 * time.Second)

					_ = stopAllDockerContainers()
					fmt.Println("Pasuing 5 seconds")
					time.Sleep(5 * time.Second)
					fmt.Println("Do you need to run './import-recipes.sh mongodb://localhost:27017' in dp-recipe-api/import-recipies ?")
					fmt.Printf("Stopping early during test number: %d", testCount)
					os.Exit(2)
				}
			} else {
				break
			}
		}

		fmt.Println("Pasuing 5 seconds")
		time.Sleep(5 * time.Second)

		err := stopAllDockerContainers()
		if err != nil {
			fmt.Printf("problem closing containers, stopping: %v\n", err)
			break
		}

		fmt.Println("Pasuing 5 seconds")
		time.Sleep(5 * time.Second)
	}
}

func startContainers() {
	cmd := exec.Command("./run-cantabular-without-sudo.sh")
	cmd.Dir = "../.."
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("compose Process: %d\n", cmd.Process.Pid)
}

func doImport() error {

	token, err := getToken()
	if err != nil {
		fmt.Println("error reading input: ", err)
		return err
	}

	// grab time before postJob to ensure we have 'time' before anything relevant to this
	// operation is put in log file for Docker container(s)
	t := time.Now()

	res, err := postJob(token)
	if err != nil {
		fmt.Println("error posting job to importAPI: ", err)
		return err
	}

	fmt.Printf("\nTrace ID (?): %s\n\n", res.ID)
	idTextFile, err := os.OpenFile(tmpFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	check(err)
	defer func() {
		cerr := idTextFile.Close()
		if cerr != nil {
			fmt.Printf("problem closing: %s : %v\n", tmpFileName, cerr)
		}
	}()

	if err := putJob(token, res); err != nil {
		fmt.Println("error putting job to importAPI: ", err)
		return err
	}

	// prefix time stamp of initiating the integration test
	// (the specific format chosen is to be compatible with the ones in the docker logs, and thus makes
	//  comparison of time's easily possible)
	_, err = fmt.Fprintf(idTextFile, "%s %s\n", t.Format("2006-01-02T15:04:05.000000000Z"), res.ID)
	check(err)

	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func postJob(token string) (*PostJobResponse, error) {
	fmt.Printf("Making request to POST import-api/jobs:")

	b := []byte(fmt.Sprintf(`{"recipe":"%s"}`, recipeID))

	fmt.Println(string(b))

	r, err := http.NewRequest("POST", importAPIHost+"/jobs", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	r.Header.Set("X-Florence-Token", token)

	res, err := httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %s", err)
	}
	defer func() {
		cerr := res.Body.Close()
		if cerr != nil {
			fmt.Printf("problem closing: response body : %v\n", cerr)
		}
	}()

	var resp PostJobResponse

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	if err := json.Unmarshal(b, &resp); err != nil {
		r := fmt.Sprintf("%d %s", res.StatusCode, string(b))
		return nil, fmt.Errorf("error unmarshalling response: '%s' response: %s\n", err, r)
	}

	fmt.Printf("Got response from POST import-api/jobs: %s\n", prettyPrint(resp))
	return &resp, nil
}

func putJob(token string, resp *PostJobResponse) error {
	fmt.Println("Making request to PUT import-api/jobs:")

	req := PutJobRequest{
		State: "submitted",
		Links: resp.Links,
	}

	b, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshalling request: %s request:\n%+v", err, req)
	}

	fmt.Println(prettyPrint(req))

	r, err := http.NewRequest("PUT", importAPIHost+"/jobs/"+resp.ID, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}

	r.Header.Set("X-Florence-Token", token)

	res, err := httpClient.Do(r)
	if err != nil {
		return fmt.Errorf("error making request: %s", err)
	}
	defer func() {
		cerr := res.Body.Close()
		if cerr != nil {
			fmt.Printf("problem closing: response body : %v\n", cerr)
		}
	}()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %s", err)
	}

	fmt.Printf("Got response from PUT import-api/jobs/%s: %d\n", resp.ID, res.StatusCode)
	fmt.Println(prettyPrint(string(b)))

	return nil
}

func prettyPrint(s interface{}) string {
	b, err := json.MarshalIndent(s, "", "  ")
	if err == nil {
		return fmt.Sprintf("%s", string(b))
	} else {
		return fmt.Sprintf("%+v", s)
	}
}

func getToken() (string, error) {
	fmt.Printf("Running get-florence-token\n")
	cmd := exec.Command("./get-florence-token.sh")
	cmd.Dir = "../.."
	var out bytes.Buffer
	cmd.Stdout = &out
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("in all caps: %q\n", out.String())
		fmt.Printf("stderr in all caps: %q\n", stdErr.String())
		return "", err
	}

	s := out.String()
	s = strings.ReplaceAll(s, "\"", "")

	fmt.Println("florence-token:", s)

	return s, nil
}

func stopAllDockerContainers() error {
	fmt.Printf("Stopping all Containers\n")
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	containersStoppedCount := 0

	for _, container := range containers {

		if strings.Contains(container.Names[0], "/cantabular-import-journey") {
			fmt.Print("Stopping container ", container.ID[:10], " ", container.Names[0], " ... \n")
			if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
				return err
			}
			containersStoppedCount++
		}
	}

	if containersStoppedCount > 0 {
		fmt.Printf("%d of %d : Containers stopped\n", containersStoppedCount, maxContainersInJob)
	} else {
		return errors.New("no containers found to stop")
	}
	return nil
}

func getContainerCount() (int, error) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return 0, err
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return 0, err
	}

	return len(containers), nil
}
