package api

import (
	"context"
	"net/http"

	dataset "github.com/ONSdigital/dp-api-clients-go/dataset"
	importapi "github.com/ONSdigital/dp-api-clients-go/importapi"
	"github.com/ONSdigital/log.go/log"
)

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
