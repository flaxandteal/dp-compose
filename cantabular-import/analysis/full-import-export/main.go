package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/importapi"
	"github.com/ONSdigital/log.go/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
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
	importAPIHost  = "http://localhost:21800"
	datasetAPIHost = "http://localhost:22000"
	recipeAPIHost  = "http://localhost:22300"
	//	recipeID               = "38542eb6-d3a6-4cc4-ba3f-32b25f23223a"
	datasetType            = "cantabular_table"
	collectionName         = "a1"
	collectionUniqueNumber = "17073b56a18b3af2f3c8220be6df9fcdaac9b5394925a9980c98bfd84ad3a003"

	client = &http.Client{}
)

const (
	idDir      = "../tmp"
	idFileName = "../tmp/id.txt"
)

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

const serviceName = "full-import-export"

// Config represents the app configuration
type Config struct {
	ImportAPIAddr              string        `envconfig:"IMPORT_API_ADDR"`
	DatasetAPIAddr             string        `envconfig:"DATASET_API_ADDR"`
	DatasetAPIMaxWorkers       int           `envconfig:"DATASET_API_MAX_WORKERS"` // maximum number of concurrent go-routines requesting items to datast api at the same time
	DatasetAPIBatchSize        int           `envconfig:"DATASET_API_BATCH_SIZE"`  // maximum size of a response by dataset api when requesting items in batches
	ShutdownTimeout            time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"                   json:"-"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
}

// NewConfig creates the config object
func NewConfig() (*Config, error) {
	cfg := Config{
		ServiceAuthToken:           "AB0A5CFA-3C55-4FA8-AACC-F98039BED0AC",
		ImportAPIAddr:              "http://localhost:21800",
		DatasetAPIAddr:             "http://localhost:22000",
		DatasetAPIMaxWorkers:       100,
		DatasetAPIBatchSize:        1000,
		ShutdownTimeout:            5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
	}
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	cfg.ServiceAuthToken = "Bearer " + cfg.ServiceAuthToken

	return &cfg, nil
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Config) String() string {
	json, _ := json.Marshal(config)
	return string(json)
}

// DatasetAPI extends the dataset api Client with json - bson mapping, specific calls, and error management
type DatasetAPI struct {
	ServiceAuthToken string
	Client           DatasetClient
	MaxWorkers       int
	BatchSize        int
}

// DatasetClient is an interface to represent methods called to action upon Dataset REST interface
type DatasetClient interface {
	GetInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID string) (m dataset.Instance, err error)
}

// errorChecker determines if an error is fatal. Only errors corresponding to http responses on the range 500+ will be considered non-fatal.
func errorChecker(ctx context.Context, tag string, err error, logData *log.Data) (isFatal bool) {
	if err == nil {
		return false
	}
	switch err.(type) {
	case *dataset.ErrInvalidDatasetAPIResponse:
		httpCode := err.(*dataset.ErrInvalidDatasetAPIResponse).Code()
		(*logData)["httpCode"] = httpCode
		if httpCode < http.StatusInternalServerError {
			isFatal = true
		}
	case *importapi.ErrInvalidAPIResponse:
		httpCode := err.(*importapi.ErrInvalidAPIResponse).Code()
		(*logData)["httpCode"] = httpCode
		if httpCode < http.StatusInternalServerError {
			isFatal = true
		}
	default:
		isFatal = true
	}
	(*logData)["is_fatal"] = isFatal
	log.Event(ctx, tag, log.ERROR, log.Error(err), *logData)
	return
}

// GetInstance asks the Dataset API for the details for instanceID
func (api *DatasetAPI) GetInstance(ctx context.Context, instanceID string) (instance dataset.Instance, isFatal bool, err error) {
	instance, err = api.Client.GetInstance(ctx, "", api.ServiceAuthToken, "", instanceID)
	isFatal = errorChecker(ctx, "GetInstance", err, &log.Data{"instanceID": instanceID})
	return
}

// logFatal is a utility method for a common failure pattern in main()
func logFatal(ctx context.Context, contextMessage string, err error, data log.Data) {
	log.Event(ctx, contextMessage, log.FATAL, log.Error(err), data)
	os.Exit(1)
}

func main() {
	fmt.Printf("Full import and export for Cantabular API's creating private and public (published) files:\n\n")
	ensureDirectoryExists(idDir)

	ctx := context.Background()
	cfg, err := NewConfig()
	if err != nil {
		logFatal(ctx, "config failed", err, nil)
	}

	token, err := getToken()
	if err != nil {
		fmt.Println("error reading input: ", err)
		os.Exit(1)
	}

	// grab time before postJob to ensure we have 'time' before anything relevant to this
	// operation is put in log file for Docker container(s)
	t := time.Now()

	recipeID, err := postCreateUniqueRecipe(token)
	if err != nil {
		fmt.Println("error creating recipe via recipe API: ", err)
		os.Exit(1)
	}

	res, err := postJob(token, recipeID)
	if err != nil {
		fmt.Println("error posting job to importAPI: ", err)
		os.Exit(1)
	}

	fmt.Printf("\nTrace ID (?): %s\n\n", res.ID)
	idTextFile, err := os.OpenFile(idFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	check(err)

	if err := putJob(token, res); err != nil {
		fmt.Println("error putting job to importAPI: ", err)
		cerr := idTextFile.Close()
		if cerr != nil {
			fmt.Printf("problem closing: %s : %v\n", idFileName, cerr)
		}
		os.Exit(1)
	}

	// prefix time stamp of initiating the integration test
	// (the specific format chosen is to be compatible with the ones in the docker logs, and thus makes
	//  comparison of time's easily possible)
	instanceID := res.Links.Instances[0].ID
	_, err = fmt.Fprintf(idTextFile, "%s %s\n", t.Format("2006-01-02T15:04:05.000000000Z"), instanceID) // instance ID
	check(err)
	cerr := idTextFile.Close()
	if cerr != nil {
		fmt.Printf("problem closing: %s : %v\n", idFileName, cerr)
	}

	// read the datasets collection and get the instance as per:
	fmt.Printf("\nThe instance'id' is: %s\n", instanceID)

	// Create wrapped datasetAPI client
	datasetAPI := &DatasetAPI{
		Client:           dataset.NewAPIClient(cfg.DatasetAPIAddr),
		ServiceAuthToken: cfg.ServiceAuthToken,
		MaxWorkers:       cfg.DatasetAPIMaxWorkers,
		BatchSize:        cfg.DatasetAPIBatchSize,
	}

	instanceFromAPI, isFatal, err := datasetAPI.GetInstance(ctx, instanceID)
	if err != nil {
		fmt.Printf("isFatal: %v\n", isFatal)
		logFatal(ctx, "GetInstance 1 failed", err, nil)
	}

	fmt.Printf("\ninstanceFromAPI: %v\n", instanceFromAPI)

	fmt.Printf("\nState: %v\n", instanceFromAPI.Version.State)

	// then check that state is : 'edition-confirmed' ... under some sort of repeat timeout

	attempts := 50

	for attempts > 0 {
		time.Sleep(100 * time.Millisecond)

		instanceFromAPI, isFatal, err = datasetAPI.GetInstance(ctx, instanceID)
		if err != nil {
			fmt.Printf("isFatal: %v\n", isFatal)
			logFatal(ctx, "GetInstance 2 failed", err, nil)
		}
		if instanceFromAPI.Version.State == "edition-confirmed" {
			// fmt.Printf("\ninstanceFromAPI: %v\n", instanceFromAPI)
			// spew.Dump(instanceFromAPI)
			fmt.Printf("Got 'edition-confirmed' after: %d milliseconds\n", 100*(51-attempts))
			break
		}
		attempts--
	}
	if attempts == 0 {
		fmt.Printf("failed to see 'edition-confirmed' after 5 seconds\n")
		os.Exit(1)
	}

	fmt.Printf("\nImport complete\n")

	fmt.Printf("\ninstance_id: %s\n", instanceFromAPI.Version.ID)
	fmt.Printf("dataset_id: %s\n", instanceFromAPI.Version.Links.Dataset.ID)
	fmt.Printf("edition: %s\n", instanceFromAPI.Version.Links.Edition.ID)
	fmt.Printf("version: %s\n", instanceFromAPI.Version.Links.Version.ID)
	// and once we have the instance and the state is as required ...

	// do the steps that produces the encrypted files ...

	// NOTE: These steps together with the parameters used have been deduced from examining docker logs
	// in the local dp-compose stack for cantabular after manually doing the florence steps to create
	// both the private and public (published) files.
	// There may be some other steps needed to fully replicate the aspects of collection database stuff,
	// but such steps do not impact the exercising of the various cantabular apis that end up achieving
	// the import, export of private and export of public (published) files.

	fmt.Printf("\nPrivate Export Step 1:\n")
	err = addDataset(token, instanceFromAPI.Version.Links.Dataset.ID, datasetType)

	if err != nil {
		fmt.Println("error doing addDataset: ", err)
		os.Exit(1)
	}

	fmt.Printf("\nPrivate Export Step 2:\n")
	err = putMetadata(token, instanceFromAPI.Version.Links.Dataset.ID)

	if err != nil {
		fmt.Println("error doing putMetadata: ", err)
		os.Exit(1)
	}

	fmt.Printf("\nPrivate Export Step 3:\n")
	err = putVersion(token,
		instanceFromAPI.Version.Links.Dataset.ID,
		instanceFromAPI.Version.Links.Edition.ID,
		instanceFromAPI.Version.Links.Version.ID,
		instanceFromAPI.Version.ID)

	if err != nil {
		fmt.Println("error doing putVersion: ", err)
		os.Exit(1)
	}

	fmt.Printf("\nPrivate Export Step 4:\n")
	err = updateInstance(token, instanceFromAPI.Version.ID) // the instance_id

	if err != nil {
		fmt.Println("error doing updateInstance: ", err)
		os.Exit(1)
	}

	fmt.Printf("\nPrivate Export Step 5:\n")
	err = putCollection(token, instanceFromAPI.Version.Links.Dataset.ID, collectionName, collectionUniqueNumber)

	if err != nil {
		fmt.Println("error doing putCollection: ", err)
		os.Exit(1)
	}

	fmt.Printf("\nPrivate Export Step 6:\n")
	err = putVersionCollection(token,
		instanceFromAPI.Version.Links.Dataset.ID,
		instanceFromAPI.Version.Links.Edition.ID,
		instanceFromAPI.Version.Links.Version.ID,
		collectionName,
		collectionUniqueNumber)

	if err != nil {
		fmt.Println("error doing putVersionCollection: ", err)
		os.Exit(1)
	}

	// then read the instance document again, looking for desired encrypted (private) file creation

	fmt.Printf("\nWaiting for 4 Private files to be created (for upt0 to 20 seconds):\n")
	attempts = 200

	for attempts > 0 {
		time.Sleep(100 * time.Millisecond)

		instanceFromAPI, isFatal, err = datasetAPI.GetInstance(ctx, instanceID)
		if err != nil {
			fmt.Printf("isFatal: %v\n", isFatal)
			logFatal(ctx, "GetInstance 3 failed", err, nil)
		}
		if instanceFromAPI.Version.Downloads["csv"].Private != "" &&
			instanceFromAPI.Version.Downloads["csvw"].Private != "" &&
			instanceFromAPI.Version.Downloads["txt"].Private != "" &&
			instanceFromAPI.Version.Downloads["xls"].Private != "" {

			fmt.Printf("\nGot all 4 private files after: %d milliseconds:\n", 100*(201-attempts))
			break
		}
		attempts--
	}
	if attempts == 0 {
		fmt.Printf("failed to see get all 4 private files after 20 seconds\nOnly got:\n")
		spew.Dump(instanceFromAPI.Version.Downloads["csv"].Private)
		spew.Dump(instanceFromAPI.Version.Downloads["csvw"].Private)
		spew.Dump(instanceFromAPI.Version.Downloads["txt"].Private)
		spew.Dump(instanceFromAPI.Version.Downloads["xls"].Private)
		os.Exit(1)
	}
	spew.Dump(instanceFromAPI.Version.Downloads["csv"].Private)
	spew.Dump(instanceFromAPI.Version.Downloads["csvw"].Private)
	spew.Dump(instanceFromAPI.Version.Downloads["txt"].Private)
	spew.Dump(instanceFromAPI.Version.Downloads["xls"].Private)

	checkFileHasContents("../../minio/data" + strings.Replace(instanceFromAPI.Version.Downloads["csv"].Private, "http://minio:9000", "", 1))
	checkFileHasContents("../../minio/data" + strings.Replace(instanceFromAPI.Version.Downloads["csvw"].Private, "http://minio:9000", "", 1))
	checkFileHasContents("../../minio/data" + strings.Replace(instanceFromAPI.Version.Downloads["txt"].Private, "http://minio:9000", "", 1))
	checkFileHasContents("../../minio/data" + strings.Replace(instanceFromAPI.Version.Downloads["xls"].Private, "http://minio:9000", "", 1))

	// do the steps that produces the unencrypted files ...

	fmt.Printf("\nPublic Export Step 7:\n")
	err = putMetadata2step7(token, instanceFromAPI.Version.Links.Dataset.ID)

	if err != nil {
		fmt.Println("error doing putMetadata2step7: ", err)
		os.Exit(1)
	}

	fmt.Printf("\nPublic Export Step 8:\n")
	err = putVersion2step8(token,
		instanceFromAPI.Version.Links.Dataset.ID,
		instanceFromAPI.Version.Links.Edition.ID,
		instanceFromAPI.Version.Links.Version.ID,
		instanceFromAPI.Version.ID)

	if err != nil {
		fmt.Println("error doing putVersion2step8: ", err)
		os.Exit(1)
	}

	fmt.Printf("\nPublic Export Step 9:\n")
	err = updateInstance2step9(token, instanceFromAPI.Version.ID) // the instance_id

	if err != nil {
		fmt.Println("error doing updateInstance2step9: ", err)
		os.Exit(1)
	}

	fmt.Printf("\nPublic Export Step 10:\n")
	err = putUpdateVersionToPublished(token,
		instanceFromAPI.Version.Links.Dataset.ID,
		instanceFromAPI.Version.Links.Edition.ID,
		instanceFromAPI.Version.Links.Version.ID)

	if err != nil {
		fmt.Println("error doing putUpdateVersionToPublished: ", err)
		os.Exit(1)
	}

	// then read the instance document again, looking for desired encrypted (private) file creation

	fmt.Printf("\nWaiting for 4 Public files to be created (for upto to 20 seconds):\n")
	attempts = 200

	for attempts > 0 {
		time.Sleep(100 * time.Millisecond)

		instanceFromAPI, isFatal, err = datasetAPI.GetInstance(ctx, instanceID)
		if err != nil {
			fmt.Printf("isFatal: %v\n", isFatal)
			logFatal(ctx, "GetInstance 3 failed", err, nil)
		}
		if instanceFromAPI.Version.Downloads["csv"].Public != "" &&
			instanceFromAPI.Version.Downloads["csvw"].Public != "" &&
			instanceFromAPI.Version.Downloads["txt"].Public != "" &&
			instanceFromAPI.Version.Downloads["xls"].Public != "" {

			fmt.Printf("\nGot all 4 public files after: %d milliseconds:\n", 100*(201-attempts))
			break
		}
		attempts--
	}
	if attempts == 0 {
		fmt.Printf("failed to see get all 4 public files after 20 seconds\nOnly got:\n")
		spew.Dump(instanceFromAPI.Version.Downloads["csv"].Public)
		spew.Dump(instanceFromAPI.Version.Downloads["csvw"].Public)
		spew.Dump(instanceFromAPI.Version.Downloads["txt"].Public)
		spew.Dump(instanceFromAPI.Version.Downloads["xls"].Public)
		os.Exit(1)
	}
	spew.Dump(instanceFromAPI.Version.Downloads["csv"].Public)
	spew.Dump(instanceFromAPI.Version.Downloads["csvw"].Public)
	spew.Dump(instanceFromAPI.Version.Downloads["txt"].Public)
	spew.Dump(instanceFromAPI.Version.Downloads["xls"].Public)

	checkFileHasContents("../../minio/data" + strings.Replace(instanceFromAPI.Version.Downloads["csv"].Public, "http://minio:9000", "", 1))
	checkFileHasContents("../../minio/data" + strings.Replace(instanceFromAPI.Version.Downloads["csvw"].Public, "http://minio:9000", "", 1))
	checkFileHasContents("../../minio/data" + strings.Replace(instanceFromAPI.Version.Downloads["txt"].Public, "http://minio:9000", "", 1))
	checkFileHasContents("../../minio/data" + strings.Replace(instanceFromAPI.Version.Downloads["xls"].Public, "http://minio:9000", "", 1))

	fmt.Printf("\nALL steps completed OK ...\n\n")
	os.Exit(0)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func ensureDirectoryExists(dirName string) {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		check(os.Mkdir(dirName, 0700))
	}
}

func checkFileHasContents(fileName string) {
	fi, err := os.Stat(fileName)
	if err != nil {
		fmt.Printf("Cant get information for: %s\n", fileName)
		os.Exit(1)
	}
	// get the size
	if fi.Size() == 0 {
		fmt.Printf("File must not have zero length: %s\n", fileName)
		os.Exit(1)
	}
}

func getToken() (string, error) {
	fmt.Printf("Running get-florence-token\n")

	cmd := exec.Command("./get-florence-token.sh") // where to get the command from
	cmd.Dir = "../.."                              // where to execute the command*/

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

	if s == "Authentication failed." {
		return s, fmt.Errorf("Failed getting token: Authentication failed")
	}

	fmt.Println("florence-token:", s)

	return s, nil
}

func postCreateUniqueRecipe(token string) (string, error) {
	fmt.Printf("\nMaking request to POST recipe-api/recipes (to create unique recipe)\n")
	uuid := uuid.New().String()

	fmt.Printf("unique uuid: %s\n", uuid)

	// create unique dataset name
	// (its a MUST for this app to be run multiple times without manually deleting a dataset of the same name)
	datasetIdName := "cantabular-example-1" + "-" + uuid

	alias := "Cantabular Example 1"
	title := "Example Cantabular Dataset City Siblings (3 mappings) And Sex"

	uri := recipeAPIHost + "/recipes"
	fmt.Printf("uri: %s\n", uri)

	body := fmt.Sprintf(`{"alias": "%s","format": "cantabular_table","id": "%s","cantabular_blob": "Example","output_instances": [{"code_lists": [{"href": "http://localhost:22400/code-lists/city-regions","id": "city","is_hierarchy": false,"name": "City"},{"href": "http://localhost:22400/code-lists/siblings_3","id": "siblings_3","is_hierarchy": false,"name": "Number Of Siblings (3 mappings)"},{"href": "http://localhost:22400/code-lists/sex","id": "sex","is_hierarchy": false,"name": "Sex"}],"dataset_id": "%s","editions": ["2021"],"title": "%s"}]}`, alias, uuid, datasetIdName, title)

	fmt.Printf("Body: %s\n", body)

	return uuid, doAPICall(token, "POST", uri, body)
}

func postJob(token, recipeID string) (*PostJobResponse, error) {
	fmt.Printf("\nMaking request to POST import-api/jobs:\n")

	b := []byte(fmt.Sprintf(`{"recipe":"%s"}`, recipeID))

	fmt.Println(string(b))

	r, err := http.NewRequest("POST", importAPIHost+"/jobs", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	r.Header.Set("X-Florence-Token", token)

	res, err := client.Do(r)
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

	fmt.Printf("\nHeader: %v\n\n", res.Header.Get("Content-Type"))

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

	// spew.Dump(resp)

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

	res, err := client.Do(r)
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

func doAPICall(token, action, uri, body string) error {
	r, err := http.NewRequest(action, uri, bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}

	r.Header.Set("X-Florence-Token", token)

	res, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("error making request: %s", err)
	}
	defer func() {
		cerr := res.Body.Close()
		if cerr != nil {
			fmt.Printf("problem closing: response body : %v\n", cerr)
		}
	}()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %s", err)
	}

	fmt.Printf("Got response from API: %d\n", res.StatusCode)
	fmt.Println(prettyPrint(string(b)))

	if res.StatusCode < 200 || res.StatusCode > 228 {
		return fmt.Errorf("Oops not OK")
	}

	return nil
}

// step 1
func addDataset(token, datasetID, datasetType string) error {
	fmt.Println("addDataset: POST /datasets/{dataset_id}:")

	uri := datasetAPIHost + "/datasets/" + datasetID
	body := fmt.Sprintf(`{"type":"%s"}`, datasetType)

	return doAPICall(token, "POST", uri, body)
}

// step 2
func putMetadata(token, datasetID string) error {
	fmt.Println("putMetadata: PUT /datasets/{dataset_id}:")

	uri := datasetAPIHost + "/datasets/" + datasetID
	body := fmt.Sprintf(`{"contacts": [{}],"id": "%s","keywords": ["a4"],"links": {"access_rights": {},"editions": {},"latest_version": {},"self": {},"taxonomy": {}},"qmi": {"href": "ons.gov.uk"},"title": "a4"}`, datasetID)

	return doAPICall(token, "PUT", uri, body)
}

// step 3
func putVersion(token, datasetID, edition, version, instanceID string) error {
	fmt.Println("putVersion: PUT /datasets/{dataset_id}/editions/{edition}/versions/{version}:")

	uri := datasetAPIHost + "/datasets/" + datasetID + "/editions/" + edition + "/versions/" + version
	body := fmt.Sprintf(`{"alerts": [],"id": "%s","links": {"dataset": {},"dimensions": {},"edition": {},"self": {}},"release_date": "2021-12-12T00:00:00.000Z","usage_notes": []}`, instanceID)

	return doAPICall(token, "PUT", uri, body)
}

// step 4
func updateInstance(token, instanceID string) error {
	fmt.Println("updateInstance: PUT /instances/{instance_id}:")

	uri := datasetAPIHost + "/instances/" + instanceID
	body := fmt.Sprintf(`{"dimensions": [{"id": "city","label": "City","links": {"code_list": {},"options": {},"version": {}},"name": "City"},{"id": "siblings_3","label": "Number of siblings (3 mappings)","links": {"code_list": {},"options": {},"version": {}},"name": "Number of siblings (3 mappings)"},{"id": "sex","label": "Sex","links": {"code_list": {},"options": {},"version": {}},"name": "Sex"}],"import_tasks": null,"last_updated": "0001-01-01T00:00:00Z"}`)

	return doAPICall(token, "PUT", uri, body)
}

// step 5
func putCollection(token, datasetID, collectionName, collectionUniqueNumber string) error {
	fmt.Println("putCollection: PUT /datasets/{dataset_id}:")

	uri := datasetAPIHost + "/datasets/" + datasetID
	body := fmt.Sprintf(`{"collection_id": "%s-%s","state": "associated"}`, collectionName, collectionUniqueNumber)

	return doAPICall(token, "PUT", uri, body)
}

// step 6
func putVersionCollection(token, datasetID, edition, version, collectionName, collectionUniqueNumber string) error {
	fmt.Println("putVersionCollection: PUT /datasets/{dataset_id}/editions/{edition}/versions/{version}:")

	uri := datasetAPIHost + "/datasets/" + datasetID + "/editions/" + edition + "/versions/" + version
	body := fmt.Sprintf(`{"collection_id": "%s-%s","state": "associated"}`, collectionName, collectionUniqueNumber)

	return doAPICall(token, "PUT", uri, body)
}

func putMetadata2step7(token, datasetID string) error {
	fmt.Println("putMetadata 2: PUT /datasets/{dataset_id}:")

	uri := datasetAPIHost + "/datasets/" + datasetID
	body := fmt.Sprintf(`{"contacts": [{}],"id": "%s","keywords": ["a4"],"links": {"access_rights": {},"editions": {},"latest_version": {},"self": {},"taxonomy": {}},"qmi": {"href": "ons.gov.uk"},"title": "a4"}`, datasetID)

	return doAPICall(token, "PUT", uri, body)
}

func putVersion2step8(token, datasetID, edition, version, instanceID string) error {
	fmt.Println("putVersion: PUT /datasets/{dataset_id}/editions/{edition}/versions/{version}:")

	uri := datasetAPIHost + "/datasets/" + datasetID + "/editions/" + edition + "/versions/" + version
	body := fmt.Sprintf(`{"alerts": [],"id": "%s","links": {"dataset": {},"dimensions": {},"edition": {},"self": {}},"release_date": "2021-12-12T00:00:00.000Z","usage_notes": []}`, instanceID)

	return doAPICall(token, "PUT", uri, body)
}

func updateInstance2step9(token, instanceID string) error {
	fmt.Println("updateInstance: PUT /instances/{instance_id}:")

	uri := datasetAPIHost + "/instances/" + instanceID
	body := fmt.Sprintf(`{"dimensions": [{"id": "city","label": "City","links": {"code_list": {},"options": {},"version": {}},"name": "City"},{"id": "siblings_3","label": "Number of siblings (3 mappings)","links": {"code_list": {},"options": {},"version": {}},"name": "Number of siblings (3 mappings)"},{"id": "sex","label": "Sex","links": {"code_list": {},"options": {},"version": {}},"name": "Sex"}],"import_tasks": null,"last_updated": "0001-01-01T00:00:00Z"}`)

	return doAPICall(token, "PUT", uri, body)
}

func putUpdateVersionToPublished(token, datasetID, edition, version string) error {
	fmt.Println("putVersion: PUT /datasets/{dataset_id}/editions/{edition}/versions/{version}:")

	uri := datasetAPIHost + "/datasets/" + datasetID + "/editions/" + edition + "/versions/" + version
	body := fmt.Sprintf(`{"state": "published"}`)

	return doAPICall(token, "PUT", uri, body)
}
