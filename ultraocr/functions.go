// Package ultraocr implements utilities to help on the UltraOCR API usage
package ultraocr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"time"

	"github.com/nuveo/ultraocr-sdk-go/ultraocr/common"
)

// NewClient Creates a client to use UltraOCR utilities.
func NewClient() client {
	return client{
		BaseURL:     common.BASE_URL,
		AuthBaseURL: common.AUTH_BASE_URL,
		Interval:    common.POOLING_INTERVAL,
		Timeout:     common.API_TIMEOUT,
		HttpClient:  http.DefaultClient,
	}
}

// SetBaseURL Changes the Client Base URL.
func (client *client) SetBaseURL(url string) {
	client.BaseURL = url
}

// SetAuthBaseURL Changes the Client Authentication Base URL.
func (client *client) SetAuthBaseURL(url string) {
	client.AuthBaseURL = url
}

// SetAuthBaseURL Changes the Client HTTP Client.
func (client *client) SetHttpClient(httpClient *http.Client) {
	client.HttpClient = httpClient
}

// SetInterval Changes the Client interval (in seconds) between requests on wait job and batch done.
func (client *client) SetInterval(interval int) {
	client.Interval = interval
}

// SetTimeout Changes the Client (timeout in seconds) on wait job and batch done.
func (client *client) SetTimeout(timeout int) {
	client.Timeout = timeout
}

// SetAutoRefresh Changes Client to auto refresh token.
func (client *client) SetAutoRefresh(clientID, clientSecret string, expires int) {
	client.ClientID = clientID
	client.ClientSecret = clientSecret
	client.Expires = expires
	client.AutoRefresh = true
	client.ExpiresAt = time.Now()
}

func (client client) request(ctx context.Context, url, method string, body io.Reader, params map[string]string) (Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return Response{}, common.ErrMountingRequest
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.Token))
	req.Header.Set("Accept", "application/json")

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	res, err := client.HttpClient.Do(req)
	if err != nil {
		return Response{}, common.ErrDoingRequest
	}

	defer res.Body.Close()

	resBody, _ := io.ReadAll(res.Body)
	return Response{
		body:   resBody,
		status: res.StatusCode,
	}, nil
}

func (client client) post(ctx context.Context, url string, body map[string]any, params map[string]string) (Response, error) {
	err := client.autoAuthenticate(ctx)
	if err != nil {
		return Response{}, err
	}

	data, err := json.Marshal(body)
	if err != nil {
		return Response{}, common.ErrParsingRequestBody
	}

	return client.request(ctx, url, http.MethodPost, bytes.NewReader(data), params)
}

func (client client) get(ctx context.Context, url string, params map[string]string) (Response, error) {
	err := client.autoAuthenticate(ctx)
	if err != nil {
		return Response{}, err
	}

	return client.request(ctx, url, http.MethodGet, nil, params)
}

func (client *client) autoAuthenticate(ctx context.Context) error {
	if client.AutoRefresh && time.Now().After(client.ExpiresAt) {
		return client.Authenticate(ctx, client.ClientID, client.ClientSecret, client.Expires)
	}

	return nil
}

func (client client) uploadFile(ctx context.Context, url string, body io.Reader) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return common.ErrMountingRequest
	}

	res, err := client.HttpClient.Do(req)
	if err != nil {
		return common.ErrDoingRequest
	}

	if res.StatusCode != 200 {
		return common.ErrInvalidStatusCode
	}

	return nil
}

// Authenticate Generates a token on UltraOCR and save the token to use on future requests.
// Requires the Client informations (ID and Secret) and the token expiration time (in minutes).
func (client *client) Authenticate(ctx context.Context, clientID, clientSecret string, expires int) error {
	url := fmt.Sprintf("%s/token", client.AuthBaseURL)
	body := map[string]any{
		"ClientID":     clientID,
		"ClientSecret": clientSecret,
		"ExpiresIn":    expires,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return common.ErrParsingRequestBody
	}

	response, err := client.request(ctx, url, http.MethodPost, bytes.NewReader(data), nil)
	if err != nil {
		return err
	}

	var res tokenResponse
	err = json.Unmarshal(response.body, &res)
	if err != nil {
		return common.ErrParsingResponse
	}

	client.Token = res.Token
	client.ExpiresAt = time.Now().Add(time.Duration(expires) * time.Minute)

	return nil
}

// GenerateSignedUrl Generates a signed url to upload the document image to be processed.
// Requires the service (document type), the resource (job or batch)
// and the required metadata and query params.
func (client *client) GenerateSignedUrl(ctx context.Context, service, resource string, metadata map[string]any, params map[string]string) (signedUrlResponse, error) {
	url := fmt.Sprintf("%s/ocr/%s/%s", client.BaseURL, resource, service)

	response, err := client.post(ctx, url, metadata, params)
	if err != nil {
		return signedUrlResponse{}, err
	}

	var res signedUrlResponse
	err = json.Unmarshal(response.body, &res)
	if err != nil {
		return signedUrlResponse{}, common.ErrParsingResponse
	}

	return res, nil
}

// UploadFileBase64 Upload a file on base64 format.
// Requires the s3 URL and the data on base64 (string).
func (client client) UploadFileBase64(ctx context.Context, url string, data string) error {
	return client.uploadFile(ctx, url, bytes.NewBufferString(data))
}

// UploadFileBase64 Upload a file given a path.
// Requires the s3 URL and the file path.
func (client client) UploadFile(ctx context.Context, url string, path string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return common.ErrReadFile
	}

	return client.uploadFile(ctx, url, bytes.NewBuffer(f))
}

// GetBatchStatus Gets the batch status. Requires the batch ID.
func (client *client) GetBatchStatus(ctx context.Context, batchID string) (batchStatusResponse, error) {
	url := fmt.Sprintf("%s/ocr/batch/status/%s", client.BaseURL, batchID)

	response, err := client.get(ctx, url, nil)
	if err != nil {
		return batchStatusResponse{}, err
	}

	var res batchStatusResponse
	err = json.Unmarshal(response.body, &res)
	if err != nil {
		return batchStatusResponse{}, common.ErrParsingResponse
	}

	return res, nil
}

// GetBatchStatus Gets the job result. Requires the batch and job ID.
func (client *client) GetJobResult(ctx context.Context, batchID, jobID string) (jobResultResponse, error) {
	url := fmt.Sprintf("%s/ocr/job/result/%s/%s", client.BaseURL, batchID, jobID)

	response, err := client.get(ctx, url, nil)
	if err != nil {
		return jobResultResponse{}, err
	}

	var res jobResultResponse
	err = json.Unmarshal(response.body, &res)
	if err != nil {
		return jobResultResponse{}, common.ErrParsingResponse
	}

	return res, nil
}

// GetJobs Gets the jobs in a time interval.
// Requires the start and end time in 2006-01-02 format.
func (client *client) GetJobs(ctx context.Context, start, end string) ([]jobResultResponse, error) {
	url := fmt.Sprintf("%s/ocr/job/results", client.BaseURL)
	params := map[string]string{
		"startDate": start,
		"endtDate":  end,
	}

	jobs := []jobResultResponse{}
	hasNextPage := true

	for hasNextPage {
		response, err := client.get(ctx, url, params)
		if err != nil {
			return nil, err
		}

		var res getJobsResponse
		err = json.Unmarshal(response.body, &res)
		if err != nil {
			return nil, common.ErrParsingResponse
		}

		jobs = append(jobs, res.Jobs...)
		params["nextPageToken"] = res.NextPageToken

		if res.NextPageToken == "" {
			hasNextPage = false
		}
	}

	return jobs, nil
}

// WaitForJobDone Waits for the job status be done or error.
// Have a timeout and an interval configured on the Client.
// Requires the batch and job ID.
func (client *client) WaitForJobDone(ctx context.Context, batchID, jobID string) (jobResultResponse, error) {
	timeout := time.Now().Add(time.Duration(client.Timeout) * time.Second)
	for {
		result, err := client.GetJobResult(ctx, batchID, jobID)
		if err != nil {
			return jobResultResponse{}, err
		}

		if result.Status == common.STATUS_DONE || result.Status == common.STATUS_ERROR {
			return result, nil
		}

		if time.Now().After(timeout) {
			return jobResultResponse{}, common.ErrTimeout
		}

		time.Sleep(time.Second * time.Duration(client.Interval))
	}
}

// WaitForBatchDone Waits for the batch status be done or error.
// Have a timeout and an interval configured on the Client.
// Requires the batch and an info if the utility will also wait the jobs to be done.
func (client *client) WaitForBatchDone(ctx context.Context, batchID string, waitJobs bool) (batchStatusResponse, error) {
	timeout := time.Now().Add(time.Duration(client.Timeout) * time.Second)
	var result batchStatusResponse

	for {
		result, err := client.GetBatchStatus(ctx, batchID)
		if err != nil {
			return batchStatusResponse{}, err
		}

		if result.Status == common.STATUS_DONE || result.Status == common.STATUS_ERROR {
			break
		}

		if time.Now().After(timeout) {
			return batchStatusResponse{}, common.ErrTimeout
		}

		time.Sleep(time.Second * time.Duration(client.Interval))
	}

	if waitJobs {
		for _, job := range result.Jobs {
			_, err := client.WaitForJobDone(ctx, batchID, job.JobID)
			if err != nil {
				return batchStatusResponse{}, err
			}
		}
	}

	return result, nil
}

// SendJobSingleStep Sends a job in single step, with 6MB body limit.
// Requires the service, the files (facematch and extra file if requested on params)
// on base64 format and the required metadata and query params.
func (client *client) SendJobSingleStep(ctx context.Context, service, file, facematchFile, extraFile string, metadata map[string]any, params map[string]string) (createdResponse, error) {
	url := fmt.Sprintf("%s/ocr/job/send/%s", client.BaseURL, service)
	body := map[string]any{
		"data":     file,
		"metadata": metadata,
	}

	if params["extra-document"] == "true" {
		body["facematch"] = facematchFile

	}

	if params["facematch"] == "true" {
		body["facematch"] = extraFile
	}

	response, err := client.post(ctx, url, body, params)
	if err != nil {
		return createdResponse{}, err
	}

	var res createdResponse
	err = json.Unmarshal(response.body, &res)
	if err != nil {
		return createdResponse{}, common.ErrParsingResponse
	}

	return res, nil
}

// SendJobBase64 Sends a job on base64 format.
// Requires the service, the files (facematch and extra file if requested on params)
// on base64 format and the required metadata and query params.
func (client *client) SendJobBase64(ctx context.Context, service, file, facematchFile, extraFile string, metadata map[string]any, params map[string]string) (createdResponse, error) {
	p := map[string]string{
		"base64": "true",
	}
	maps.Copy(p, params)

	response, err := client.GenerateSignedUrl(ctx, service, common.RESOURCE_JOB, metadata, p)
	if err != nil {
		return createdResponse{}, err
	}

	urls := response.URLs
	err = client.UploadFileBase64(ctx, urls["document"], file)
	if err != nil {
		return createdResponse{}, err
	}

	if p["facematch"] == "true" {
		err = client.UploadFileBase64(ctx, urls["selfie"], extraFile)
		if err != nil {
			return createdResponse{}, err
		}
	}

	if p["extra-document"] == "true" {
		err = client.UploadFileBase64(ctx, urls["extra_document"], facematchFile)
		if err != nil {
			return createdResponse{}, err
		}
	}

	return createdResponse{
		Id:        response.Id,
		StatusURL: response.StatusURL,
	}, nil
}

// SendJob Sends a job.
// Requires the service, the files (facematch and extra file if requested on params) paths
// and the required metadata and query params.
func (client *client) SendJob(ctx context.Context, service, filePath, facematchFilePath, extraFilePath string, metadata map[string]any, params map[string]string) (createdResponse, error) {
	response, err := client.GenerateSignedUrl(ctx, service, common.RESOURCE_JOB, metadata, params)
	if err != nil {
		return createdResponse{}, err
	}

	urls := response.URLs
	err = client.UploadFile(ctx, urls["document"], filePath)
	if err != nil {
		return createdResponse{}, err
	}

	if params["facematch"] == "true" {
		err = client.UploadFile(ctx, urls["selfie"], extraFilePath)
		if err != nil {
			return createdResponse{}, err
		}
	}

	if params["extra-document"] == "true" {
		err = client.UploadFile(ctx, urls["extra_document"], facematchFilePath)
		if err != nil {
			return createdResponse{}, err
		}
	}

	return createdResponse{
		Id:        response.Id,
		StatusURL: response.StatusURL,
	}, nil
}

// SendBatchBase64 Sends a batch on base64 format.
// Requires the service, the file on base64 format and the required metadata and query params.
func (client *client) SendBatchBase64(ctx context.Context, service, file string, metadata map[string]any, params map[string]string) (createdResponse, error) {
	p := map[string]string{
		"base64": "true",
	}
	maps.Copy(p, params)

	response, err := client.GenerateSignedUrl(ctx, service, common.RESOURCE_BATCH, metadata, p)
	if err != nil {
		return createdResponse{}, err
	}

	urls := response.URLs
	err = client.UploadFileBase64(ctx, urls["document"], file)
	if err != nil {
		return createdResponse{}, err
	}

	return createdResponse{
		Id:        response.Id,
		StatusURL: response.StatusURL,
	}, nil
}

// SendBatch Sends a batch.
// Requires the service, the file path and the required metadata and query params.
func (client *client) SendBatch(ctx context.Context, service, filePath string, metadata map[string]any, params map[string]string) (createdResponse, error) {
	response, err := client.GenerateSignedUrl(ctx, service, common.RESOURCE_BATCH, metadata, params)
	if err != nil {
		return createdResponse{}, err
	}

	urls := response.URLs
	err = client.UploadFile(ctx, urls["document"], filePath)
	if err != nil {
		return createdResponse{}, err
	}

	return createdResponse{
		Id:        response.Id,
		StatusURL: response.StatusURL,
	}, nil
}

// CreateAndWaitJob Creates and wait a job to be done.
// Have a timeout and an interval configured on the Client.
// Requires the service, files paths and required metadata and query params.
func (client *client) CreateAndWaitJob(ctx context.Context, service, filePath, facematchFilePath, extraFilePath string, metadata map[string]any, params map[string]string) (jobResultResponse, error) {
	response, err := client.SendJob(ctx, service, filePath, facematchFilePath, extraFilePath, metadata, params)
	if err != nil {
		return jobResultResponse{}, err
	}

	jobID := response.Id
	return client.WaitForJobDone(ctx, jobID, jobID)
}

// CreateAndWaitJob Creates and wait a batch to be done.
// Have a timeout and an interval configured on the Client.
// Requires the service, file path and required metadata and query params.
func (client *client) CreateAndWaitBatch(ctx context.Context, service, filePath string, metadata map[string]any, params map[string]string, waitJobs bool) (batchStatusResponse, error) {
	response, err := client.SendBatch(ctx, service, filePath, metadata, params)
	if err != nil {
		return batchStatusResponse{}, err
	}

	return client.WaitForBatchDone(ctx, response.Id, waitJobs)
}
