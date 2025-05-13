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
	Expires   int               `json:"exp"`
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

type BatchInfoResponse struct {
	ValidationID   string `json:"validation_id,omitempty"`
	BatchID        string `json:"batch_id,omitempty"`
	ClientID       string `json:"client_id,omitempty"`
	CompanyID      string `json:"company_id,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	Error          string `json:"error,omitempty"`
	Service        string `json:"service"`
	Source         string `json:"source,omitempty"`
	Status         string `json:"status"`
	TotalJobs      int    `json:"total_jobs,omitempty"`
	TotalProcessed int    `json:"total_processed,omitempty"`
}

type JobInfoResponse struct {
	ValidationID     string `json:"validation_id,omitempty"`
	ClientID         string `json:"client_id,omitempty"`
	CompanyID        string `json:"company_id,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	Error            string `json:"error,omitempty"`
	FinishedAt       string `json:"finished_at,omitempty"`
	JobID            string `json:"job_id,omitempty"`
	Result           Result `json:"result,omitempty"`
	Service          string `json:"service"`
	Source           string `json:"source,omitempty"`
	Status           string `json:"status"`
	ValidationStatus string `json:"validation_status,omitempty"`
	ClientData       any    `json:"client_data,omitempty"`
	Metadata         any    `json:"metadata,omitempty"`
	Validation       any    `json:"validation,omitempty"`
}

type BatchResultJob struct {
	JobKSUID         string `json:"job_ksuid"`
	Status           string `json:"status"`
	Service          string `json:"service"`
	Error            string `json:"error,omitempty"`
	Result           Result `json:"result,omitempty"`
	Filename         string `json:"filename,omitempty"`
	ClientData       any    `json:"client_data,omitempty"`
	Validation       any    `json:"validation,omitempty"`
	ValidationStatus string `json:"validation_status,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
}

type BatchResultStorageResponse struct {
	Url string `json:"url"`
	Exp string `json:"exp"`
}
