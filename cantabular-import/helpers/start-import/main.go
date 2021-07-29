package main

import (
	"net/http"
	"bufio"
	"io/ioutil"
	"io"
	"os"
	"fmt"
	"encoding/json"
	"bytes"
)

type PutJobRequest struct{
	Links Links  `json:"links"`
	State string `json:"state"`
}

type PostJobResponse struct{
	ID string   `json:"id"`
	Links Links `json:"links"`
}

type Links struct{
	Instances []Link `json:"instances"`
	Self Link        `json:"self"`
}

type Link struct{
	ID   string `json:"id"`
	HRef string `json:"href"`
}

var (
	importAPIHost = "http://localhost:21800"
	recipeID =      "38542eb6-d3a6-4cc4-ba3f-32b25f23223a"

	client = &http.Client{}
)

func main(){
	token, err := readInput()
	if err != nil{
		fmt.Println("error reading input: ", err)
	}

	res, err := postJob(token)
	if err != nil{
		fmt.Println("error posting job to importAPI: ", err)
		os.Exit(1)
	}

	if err := putJob(token, res); err != nil{
		fmt.Println("error putting job to importAPI: ", err)
		os.Exit(1)
	}
}

func readInput() (string, error){
	rdr := bufio.NewReader(os.Stdin)

	s, err := rdr.ReadString('\n')
	if err != nil && err != io.EOF{
		return "", fmt.Errorf("failed to read from stdin: %s", err)
	}

	fmt.Println("florence-token:", s)

	return s, nil
}

func postJob(token string) (*PostJobResponse, error){
	fmt.Printf("Making request to POST import-api/jobs:")

	b := []byte(fmt.Sprintf(`{"recipe":"%s"}`, recipeID))

	fmt.Println(string(b))

	r, err :=  http.NewRequest("POST", importAPIHost + "/jobs", bytes.NewReader(b))
	if err != nil{
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	r.Header.Set("X-Florence-Token", token)

	res, err := client.Do(r)
	if err != nil{
		return nil, fmt.Errorf("error performing request: %s", err)
	}
	defer res.Body.Close()

	var resp PostJobResponse

	b, err = ioutil.ReadAll(res.Body)
	if err != nil{
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	if err := json.Unmarshal(b, &resp); err != nil{
		r := fmt.Sprintf("%d %s", res.StatusCode, string(b))
		return nil, fmt.Errorf("error unmarshalling response: '%s' response: %s\n", err, r)
	}

	fmt.Printf("Got response from POST import-api/jobs: %s\n", prettyPrint(resp))
	return &resp, nil
}

func putJob(token string, resp *PostJobResponse) error{
	fmt.Println("Making request to PUT import-api/jobs:")

	req := PutJobRequest{
		State: "submitted",
		Links: resp.Links,
	}

	b, err := json.Marshal(req)
	if err != nil{
		return fmt.Errorf("error marshalling request: %s request:\n%+v", err, req)
	}

	fmt.Println(prettyPrint(req))

	r, err :=  http.NewRequest("PUT", importAPIHost + "/jobs/" + resp.ID, bytes.NewReader(b))
	if err != nil{
		return fmt.Errorf("error creating request: %s", err)
	}

	r.Header.Set("X-Florence-Token", token)

	res, err := client.Do(r)
	if err != nil{
		return fmt.Errorf("error making request: %s", err)
	}
	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil{
		return fmt.Errorf("error reading response body: %s", err)
	}

	fmt.Printf("Got response from PUT import-api/jobs/%s: %d\n", resp.ID, res.StatusCode)
	fmt.Println(prettyPrint(string(b)))

	return nil
}

func prettyPrint(s interface{}) string {
	b, err := json.MarshalIndent(s, "", "  ")
	if err == nil{
		return fmt.Sprintf("%s", string(b))
	} else {
		return fmt.Sprintf("%+v", s)
	}
}
