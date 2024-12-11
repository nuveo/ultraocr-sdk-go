// Package ultraocr implements utilities to help on the UltraOCR API usage.
package ultraocr

import (
	"net/http"
	"time"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	BaseURL      string
	AuthBaseURL  string
	Token        string
	ClientID     string
	ClientSecret string
	AutoRefresh  bool
	Expires      int
	Timeout      int
	Interval     int
	ExpiresAt    time.Time
	HttpClient   HttpClient
}

type Response struct {
	body   []byte
	status int
}

type tokenResponse struct {
	Token string `json:"token"`
}

type SignedUrlResponse struct {
	Expires   string            `json:"exp"`
	Id        string            `json:"id"`
	StatusURL string            `json:"status_url"`
	URLs      map[string]string `json:"urls"`
}

type CreatedResponse struct {
	Id        string `json:"id"`
	StatusURL string `json:"status_url"`
}

type BatchStatusJobs struct {
	JobID     string `json:"job_ksuid"`
	CreatedAt string `json:"created_at"`
	ResultURL string `json:"result_url"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

type BatchStatusResponse struct {
	BatchID   string            `json:"batch_ksuid"`
	CreatedAt string            `json:"created_at"`
	Service   string            `json:"service"`
	Status    string            `json:"status"`
	Error     string            `json:"error,omitempty"`
	Jobs      []BatchStatusJobs `json:"jobs"`
}

type Result struct {
	Document interface{} `json:"Document,omitempty"`
	Quantity int         `json:"Quantity,omitempty"`
	Time     string      `json:"Time,omitempty"`
}

type JobResultResponse struct {
	Result           Result      `json:"result,omitempty"`
	JobID            string      `json:"job_ksuid"`
	CreatedAt        string      `json:"created_at"`
	Service          string      `json:"service"`
	Status           string      `json:"status"`
	Error            string      `json:"error,omitempty"`
	ProcessTime      string      `json:"process_time,omitempty"`
	Filename         string      `json:"filename,omitempty"`
	ValidationStatus string      `json:"validation_status,omitempty"`
	ClientData       interface{} `json:"client_data,omitempty"`
	Validation       interface{} `json:"validation,omitempty"`
}

type GetJobsResponse struct {
	Jobs          []JobResultResponse `json:"jobs"`
	NextPageToken string              `json:"nextPageToken"`
}
