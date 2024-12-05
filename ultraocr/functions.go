package ultraocr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nuveo/ultraocr-sdk-go/ultraocr/common"
)

func NewClient() client {
	return client{
		BaseURL:     common.BASE_URL,
		AuthBaseURL: common.AUTH_BASE_URL,
		Interval:    common.POOLING_INTERVAL,
		Timeout:     common.API_TIMEOUT,
		HttpClient:  http.DefaultClient,
	}
}

func (client *client) SetBaseURL(url string) {
	client.BaseURL = url
}

func (client *client) SetAuthBaseURL(url string) {
	client.AuthBaseURL = url
}

func (client *client) SetHttpClient(httpClient *http.Client) {
	client.HttpClient = httpClient
}

func (client *client) SetInterval(interval int) {
	client.Interval = interval
}

func (client *client) SetTimeout(timeout int) {
	client.Timeout = timeout
}

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
		return Response{}, err
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
		return Response{}, err
	}

	defer res.Body.Close()

	resBody, _ := io.ReadAll(res.Body)
	return Response{
		body:   resBody,
		status: res.StatusCode,
	}, nil
}

func (client *client) Authenticate(ctx context.Context, clientID, clientSecret string, expires int) error {
	url := fmt.Sprintf("%s/token", client.AuthBaseURL)
	body := map[string]any{
		"ClientID":     clientID,
		"ClientSecret": clientSecret,
		"ExpiresIn":    expires,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	response, err := client.request(ctx, url, http.MethodPost, bytes.NewReader(data), nil)
	if err != nil {
		return err
	}

	var res tokenResponse
	err = json.Unmarshal(response.body, &res)
	if err != nil {
		return err
	}

	client.Token = res.Token
	client.ExpiresAt = time.Now().Add(time.Duration(expires) * time.Minute)

	return nil
}

func (client *client) autoAuthenticate(ctx context.Context) error {
	if client.AutoRefresh && time.Now().After(client.ExpiresAt) {
		return client.Authenticate(ctx, client.ClientID, client.ClientSecret, client.Expires)
	}

	return nil
}

func (client client) post(ctx context.Context, url string, body map[string]any, params map[string]string) (Response, error) {
	err := client.autoAuthenticate(ctx)
	if err != nil {
		return Response{}, err
	}

	data, err := json.Marshal(body)
	if err != nil {
		return Response{}, err
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

func (client *client) GenerateSignedUrl(ctx context.Context, service, resource string, metadata map[string]any, params map[string]string) (signedUrlResponse, error) {
	url := fmt.Sprintf("%s/ocr/%s/%s", client.BaseURL, resource, service)

	response, err := client.post(ctx, url, metadata, params)
	if err != nil {
		return signedUrlResponse{}, err
	}

	var res signedUrlResponse
	err = json.Unmarshal(response.body, &res)

	return res, err
}

func (client *client) GetBatchStatus(ctx context.Context, batchID string) (batchStatusResponse, error) {
	url := fmt.Sprintf("%s/ocr/batch/status/%s", client.BaseURL, batchID)

	response, err := client.get(ctx, url, nil)
	if err != nil {
		return batchStatusResponse{}, err
	}

	var res batchStatusResponse
	err = json.Unmarshal(response.body, &res)

	return res, err
}

func (client *client) GetJobResult(ctx context.Context, batchID, jobID string) (jobResultResponse, error) {
	url := fmt.Sprintf("%s/ocr/job/result/%s/%s", client.BaseURL, batchID, jobID)

	response, err := client.get(ctx, url, nil)
	if err != nil {
		return jobResultResponse{}, err
	}

	var res jobResultResponse
	err = json.Unmarshal(response.body, &res)

	return res, err
}
