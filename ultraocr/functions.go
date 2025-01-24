// Package ultraocr implements utilities to help on the UltraOCR API usage.
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
func NewClient() Client {
	return Client{
		BaseURL:     common.BASE_URL,
		AuthBaseURL: common.AUTH_BASE_URL,
		Interval:    common.POOLING_INTERVAL,
		Timeout:     common.API_TIMEOUT,
		HttpClient:  http.DefaultClient,
	}
}

// SetBaseURL Changes the Client Base URL.
func (client *Client) SetBaseURL(url string) {
	client.BaseURL = url
}

// SetAuthBaseURL Changes the Client Authentication Base URL.
func (client *Client) SetAuthBaseURL(url string) {
	client.AuthBaseURL = url
}

// SetHttpClient Changes the Client HTTP Client.
func (client *Client) SetHttpClient(httpClient HttpClient) {
	client.HttpClient = httpClient
}

// SetInterval Changes the Client interval (in seconds) between requests on wait job and batch done.
func (client *Client) SetInterval(interval int) {
	client.Interval = interval
}

// SetTimeout Changes the Client (timeout in seconds) on wait job and batch done.
func (client *Client) SetTimeout(timeout int) {
	client.Timeout = timeout
}

// SetAutoRefresh Changes Client to auto refresh token.
func (client *Client) SetAutoRefresh(clientID, clientSecret string, expires int) {
	client.ClientID = clientID
	client.ClientSecret = clientSecret
	client.Expires = expires
	client.AutoRefresh = true
	client.ExpiresAt = time.Now()
}

func (client Client) request(
	ctx context.Context,
	url,
	method string,
	body io.Reader,
	params map[string]string,
) (Response, error) {
	err := client.autoAuthenticate(ctx)
	if err != nil {
		return Response{}, err
	}

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

func (client Client) post(
	ctx context.Context,
	url string,
	body any,
	params map[string]string,
) (Response, error) {
	if isNil(body) {
		data, err := json.Marshal(body)
		if err != nil {
			return Response{}, common.ErrParsingRequestBody
		}

		return client.request(ctx, url, http.MethodPost, bytes.NewReader(data), params)
	}

	return client.request(ctx, url, http.MethodPost, nil, params)
}

func (client Client) get(ctx context.Context, url string, params map[string]string) (Response, error) {
	return client.request(ctx, url, http.MethodGet, nil, params)
}

func (client *Client) autoAuthenticate(ctx context.Context) error {
	if client.AutoRefresh && time.Now().After(client.ExpiresAt) {
		return client.Authenticate(ctx, client.ClientID, client.ClientSecret, client.Expires)
	}

	return nil
}

func (client Client) uploadFile(ctx context.Context, url string, body io.Reader) error {
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
func (client *Client) Authenticate(ctx context.Context, clientID, clientSecret string, expires int) error {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return common.ErrMountingRequest
	}
	req.Header.Set("Accept", "application/json")

	response, err := client.HttpClient.Do(req)
	if err != nil {
		return common.ErrDoingRequest
	}

	defer response.Body.Close()

	resBody, _ := io.ReadAll(response.Body)
	if response.StatusCode != 200 {
		return common.ErrInvalidStatusCode
	}

	var res tokenResponse
	err = json.Unmarshal(resBody, &res)
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
func (client *Client) GenerateSignedUrl(
	ctx context.Context,
	service,
	resource string,
	metadata any,
	params map[string]string,
) (SignedUrlResponse, error) {
	url := fmt.Sprintf("%s/ocr/%s/%s", client.BaseURL, resource, service)

	response, err := client.post(ctx, url, metadata, params)
	if err != nil {
		return SignedUrlResponse{}, err
	}

	if response.status != 200 {
		return SignedUrlResponse{}, common.ErrInvalidStatusCode
	}

	var res SignedUrlResponse
	err = json.Unmarshal(response.body, &res)
	if err != nil {
		return SignedUrlResponse{}, common.ErrParsingResponse
	}

	return res, nil
}

// UploadFileBase64 Upload a file on base64 format.
// Requires the s3 URL and the data on base64 (string).
func (client Client) UploadFileBase64(ctx context.Context, url string, data string) error {
	return client.uploadFile(ctx, url, bytes.NewBufferString(data))
}

// UploadFileBase64 Upload a file given a path.
// Requires the s3 URL and the file path.
func (client Client) UploadFile(ctx context.Context, url string, path string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return common.ErrReadFile
	}

	return client.uploadFile(ctx, url, bytes.NewBuffer(f))
}

// GetBatchStatus Gets the batch status. Requires the batch ID.
func (client *Client) GetBatchStatus(ctx context.Context, ID string) (BatchStatusResponse, error) {
	url := fmt.Sprintf("%s/ocr/batch/status/%s", client.BaseURL, ID)

	response, err := client.get(ctx, url, nil)
	if err != nil {
		return BatchStatusResponse{}, err
	}

	if response.status != 200 {
		return BatchStatusResponse{}, common.ErrInvalidStatusCode
	}

	var res BatchStatusResponse
	err = json.Unmarshal(response.body, &res)
	if err != nil {
		return BatchStatusResponse{}, common.ErrParsingResponse
	}

	return res, nil
}

// GetBatchStatus Gets the job result. Requires the batch and job ID.
func (client *Client) GetJobResult(ctx context.Context, batchID, jobID string) (JobResultResponse, error) {
	url := fmt.Sprintf("%s/ocr/job/result/%s/%s", client.BaseURL, batchID, jobID)

	response, err := client.get(ctx, url, nil)
	if err != nil {
		return JobResultResponse{}, err
	}

	if response.status != 200 {
		return JobResultResponse{}, common.ErrInvalidStatusCode
	}

	var res JobResultResponse
	err = json.Unmarshal(response.body, &res)
	if err != nil {
		return JobResultResponse{}, common.ErrParsingResponse
	}

	return res, nil
}

// GetJobs Gets the jobs in a time interval.
// Requires the start and end time in 2006-01-02 format.
func (client *Client) GetJobs(ctx context.Context, start, end string) ([]JobResultResponse, error) {
	url := fmt.Sprintf("%s/ocr/job/results", client.BaseURL)
	params := map[string]string{
		"startDate": start,
		"endtDate":  end,
	}

	jobs := []JobResultResponse{}
	hasNextPage := true

	for hasNextPage {
		response, err := client.get(ctx, url, params)
		if err != nil {
			return nil, err
		}

		if response.status != 200 {
			return nil, common.ErrInvalidStatusCode
		}

		var res GetJobsResponse
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

// SendJobSingleStep Sends a job in single step, with 6MB body limit.
// Requires the service, the files (facematch and extra file if requested on params)
// on base64 format and the required metadata and query params.
func (client *Client) SendJobSingleStep(
	ctx context.Context,
	service,
	file,
	facematchFile,
	extraFile string,
	metadata map[string]any,
	params map[string]string,
) (CreatedResponse, error) {
	url := fmt.Sprintf("%s/ocr/job/send/%s", client.BaseURL, service)
	body := map[string]any{
		"data":     file,
		"metadata": metadata,
	}

	if params[common.KEY_EXTRA] == common.FLAG_TRUE {
		body["extra"] = extraFile

	}

	if params[common.KEY_FACEMATCH] == common.FLAG_TRUE {
		body[common.KEY_FACEMATCH] = facematchFile
	}

	response, err := client.post(ctx, url, body, params)
	if err != nil {
		return CreatedResponse{}, err
	}

	if response.status != 200 {
		return CreatedResponse{}, common.ErrInvalidStatusCode
	}

	var res CreatedResponse
	err = json.Unmarshal(response.body, &res)
	if err != nil {
		return CreatedResponse{}, common.ErrParsingResponse
	}

	return res, nil
}

// SendJobBase64 Sends a job on base64 format.
// Requires the service, the files (facematch and extra file if requested on params)
// on base64 format and the required metadata and query params.
func (client *Client) SendJobBase64(ctx context.Context,
	service,
	file,
	facematchFile,
	extraFile string,
	metadata map[string]any,
	params map[string]string,
) (CreatedResponse, error) {
	p := map[string]string{
		"base64": common.FLAG_TRUE,
	}
	maps.Copy(p, params)

	response, err := client.GenerateSignedUrl(ctx, service, common.RESOURCE_JOB, metadata, p)
	if err != nil {
		return CreatedResponse{}, err
	}

	urls := response.URLs
	err = client.UploadFileBase64(ctx, urls["document"], file)
	if err != nil {
		return CreatedResponse{}, err
	}

	if p[common.KEY_FACEMATCH] == common.FLAG_TRUE {
		err = client.UploadFileBase64(ctx, urls["selfie"], facematchFile)
		if err != nil {
			return CreatedResponse{}, err
		}
	}

	if p[common.KEY_EXTRA] == common.FLAG_TRUE {
		err = client.UploadFileBase64(ctx, urls["extra_document"], extraFile)
		if err != nil {
			return CreatedResponse{}, err
		}
	}

	return CreatedResponse{
		Id:        response.Id,
		StatusURL: response.StatusURL,
	}, nil
}

// SendJob Sends a job.
// Requires the service, the files (facematch and extra file if requested on params) paths
// and the required metadata and query params.
func (client *Client) SendJob(ctx context.Context,
	service,
	filePath,
	facematchFilePath,
	extraFilePath string,
	metadata map[string]any,
	params map[string]string,
) (CreatedResponse, error) {
	response, err := client.GenerateSignedUrl(ctx, service, common.RESOURCE_JOB, metadata, params)
	if err != nil {
		return CreatedResponse{}, err
	}

	urls := response.URLs
	err = client.UploadFile(ctx, urls["document"], filePath)
	if err != nil {
		return CreatedResponse{}, err
	}

	if params[common.KEY_FACEMATCH] == common.FLAG_TRUE {
		err = client.UploadFile(ctx, urls["selfie"], facematchFilePath)
		if err != nil {
			return CreatedResponse{}, err
		}
	}

	if params[common.KEY_EXTRA] == common.FLAG_TRUE {
		err = client.UploadFile(ctx, urls["extra_document"], extraFilePath)
		if err != nil {
			return CreatedResponse{}, err
		}
	}

	return CreatedResponse{
		Id:        response.Id,
		StatusURL: response.StatusURL,
	}, nil
}

// SendBatchBase64 Sends a batch on base64 format.
// Requires the service, the file on base64 format and the required metadata and query params.
func (client *Client) SendBatchBase64(ctx context.Context,
	service,
	file string,
	metadata []map[string]any,
	params map[string]string,
) (CreatedResponse, error) {
	p := map[string]string{
		"base64": common.FLAG_TRUE,
	}
	maps.Copy(p, params)

	response, err := client.GenerateSignedUrl(ctx, service, common.RESOURCE_BATCH, metadata, p)
	if err != nil {
		return CreatedResponse{}, err
	}

	urls := response.URLs
	err = client.UploadFileBase64(ctx, urls["document"], file)
	if err != nil {
		return CreatedResponse{}, err
	}

	return CreatedResponse{
		Id:        response.Id,
		StatusURL: response.StatusURL,
	}, nil
}

// SendBatch Sends a batch.
// Requires the service, the file path and the required metadata and query params.
func (client *Client) SendBatch(ctx context.Context,
	service,
	filePath string,
	metadata []map[string]any,
	params map[string]string,
) (CreatedResponse, error) {
	response, err := client.GenerateSignedUrl(ctx, service, common.RESOURCE_BATCH, metadata, params)
	if err != nil {
		return CreatedResponse{}, err
	}

	urls := response.URLs
	err = client.UploadFile(ctx, urls["document"], filePath)
	if err != nil {
		return CreatedResponse{}, err
	}

	return CreatedResponse{
		Id:        response.Id,
		StatusURL: response.StatusURL,
	}, nil
}

// WaitForJobDone Waits for the job status be done or error.
// Have a timeout and an interval configured on the Client.
// Requires the batch and job ID.
func (client *Client) WaitForJobDone(ctx context.Context, batchID, jobID string) (JobResultResponse, error) {
	timeout := time.Now().Add(time.Duration(client.Timeout) * time.Second)
	for {
		result, err := client.GetJobResult(ctx, batchID, jobID)
		if err != nil {
			return JobResultResponse{}, err
		}

		if result.Status == common.STATUS_DONE || result.Status == common.STATUS_ERROR {
			return result, nil
		}

		if time.Now().After(timeout) {
			return JobResultResponse{}, common.ErrTimeout
		}

		time.Sleep(time.Second * time.Duration(client.Interval))
	}
}

// WaitForBatchDone Waits for the batch status be done or error.
// Have a timeout and an interval configured on the Client.
// Requires the batch and an info if the utility will also wait the jobs to be done.
func (client *Client) WaitForBatchDone(ctx context.Context, ID string, waitJobs bool) (BatchStatusResponse, error) {
	timeout := time.Now().Add(time.Duration(client.Timeout) * time.Second)
	var result BatchStatusResponse
	var err error

	for {
		result, err = client.GetBatchStatus(ctx, ID)
		if err != nil {
			return BatchStatusResponse{}, err
		}

		if result.Status == common.STATUS_DONE || result.Status == common.STATUS_ERROR {
			break
		}

		if time.Now().After(timeout) {
			return BatchStatusResponse{}, common.ErrTimeout
		}

		time.Sleep(time.Second * time.Duration(client.Interval))
	}

	if waitJobs {
		for _, job := range result.Jobs {
			_, err := client.WaitForJobDone(ctx, ID, job.JobID)
			if err != nil {
				return BatchStatusResponse{}, err
			}
		}
	}

	return result, nil
}

// CreateAndWaitJob Creates and wait a job to be done.
// Have a timeout and an interval configured on the Client.
// Requires the service, files paths and required metadata and query params.
func (client *Client) CreateAndWaitJob(ctx context.Context,
	service,
	filePath,
	facematchFilePath,
	extraFilePath string,
	metadata map[string]any,
	params map[string]string,
) (JobResultResponse, error) {
	response, err := client.SendJob(ctx, service, filePath, facematchFilePath, extraFilePath, metadata, params)
	if err != nil {
		return JobResultResponse{}, err
	}

	jobID := response.Id
	return client.WaitForJobDone(ctx, jobID, jobID)
}

// CreateAndWaitJob Creates and wait a batch to be done.
// Have a timeout and an interval configured on the Client.
// Requires the service, file path and required metadata and query params.
func (client *Client) CreateAndWaitBatch(ctx context.Context,
	service,
	filePath string,
	metadata []map[string]any,
	params map[string]string,
	waitJobs bool,
) (BatchStatusResponse, error) {
	response, err := client.SendBatch(ctx, service, filePath, metadata, params)
	if err != nil {
		return BatchStatusResponse{}, err
	}

	return client.WaitForBatchDone(ctx, response.Id, waitJobs)
}
