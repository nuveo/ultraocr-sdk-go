// Package ultraocr implements utilities to help on the UltraOCR API usage.
package ultraocr

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nuveo/ultraocr-sdk-go/ultraocr/common"
)

type ClientMock struct {
	MockDo func(req *http.Request) (*http.Response, error)
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	return c.MockDo(req)
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name string
		want Client
	}{
		{
			name: "success",
			want: Client{
				BaseURL:     common.BASE_URL,
				AuthBaseURL: common.AUTH_BASE_URL,
				Interval:    common.POOLING_INTERVAL,
				Timeout:     common.API_TIMEOUT,
				HttpClient:  http.DefaultClient,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSets(t *testing.T) {
	t.Run("test sets", func(t *testing.T) {
		c := NewClient()

		c.SetBaseURL("url")
		want := Client{
			BaseURL:     "url",
			AuthBaseURL: common.AUTH_BASE_URL,
			Interval:    common.POOLING_INTERVAL,
			Timeout:     common.API_TIMEOUT,
			HttpClient:  http.DefaultClient,
		}
		if !reflect.DeepEqual(c, want) {
			t.Errorf("client = %v, want %v", c, want)
		}

		c.SetAuthBaseURL("url")
		want = Client{
			BaseURL:     "url",
			AuthBaseURL: "url",
			Interval:    common.POOLING_INTERVAL,
			Timeout:     common.API_TIMEOUT,
			HttpClient:  http.DefaultClient,
		}
		if !reflect.DeepEqual(c, want) {
			t.Errorf("client = %v, want %v", c, want)
		}

		c.SetInterval(3)
		want = Client{
			BaseURL:     "url",
			AuthBaseURL: "url",
			Interval:    3,
			Timeout:     common.API_TIMEOUT,
			HttpClient:  http.DefaultClient,
		}
		if !reflect.DeepEqual(c, want) {
			t.Errorf("client = %v, want %v", c, want)
		}

		c.SetTimeout(10)
		want = Client{
			BaseURL:     "url",
			AuthBaseURL: "url",
			Interval:    3,
			Timeout:     10,
			HttpClient:  http.DefaultClient,
		}
		if !reflect.DeepEqual(c, want) {
			t.Errorf("client = %v, want %v", c, want)
		}

		c.SetHttpClient(&http.Client{
			Timeout: 20,
		})
		want = Client{
			BaseURL:     "url",
			AuthBaseURL: "url",
			Interval:    3,
			Timeout:     10,
			HttpClient: &http.Client{
				Timeout: 20,
			},
		}
		if !reflect.DeepEqual(c, want) {
			t.Errorf("client = %v, want %v", c, want)
		}

		c.SetAutoRefresh("id", "secret", 10)
		want = Client{
			BaseURL:     "url",
			AuthBaseURL: "url",
			Interval:    3,
			Timeout:     10,
			HttpClient: &http.Client{
				Timeout: 20,
			},
			ClientID:     "id",
			ClientSecret: "secret",
			Expires:      10,
			AutoRefresh:  true,
		}
		if c.ClientID != want.ClientID {
			t.Errorf("client = %v, want %v", c.ClientID, want.ClientID)
		}
		if c.ClientSecret != want.ClientSecret {
			t.Errorf("client = %v, want %v", c.ClientSecret, want.ClientSecret)
		}
		if c.Expires != want.Expires {
			t.Errorf("client = %v, want %v", c.Expires, want.Expires)
		}
		if c.AutoRefresh != want.AutoRefresh {
			t.Errorf("client = %v, want %v", c.AutoRefresh, want.AutoRefresh)
		}
	})
}

func TestRequest(t *testing.T) {
	type fields struct {
		HttpClient ClientMock
	}
	type args struct {
		ctx    context.Context
		url    string
		method string
		body   io.Reader
		params map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Response
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			args: args{
				params: map[string]string{
					"some": "param",
				},
				ctx: context.Background(),
			},
			want: Response{
				body:   []byte{},
				status: 200,
			},
		},
		{
			name: "fail to do request",
			fields: fields{
				HttpClient: ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name:    "fail to mount request",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				HttpClient: &tt.fields.HttpClient,
			}
			got, err := client.request(tt.args.ctx, tt.args.url, tt.args.method, tt.args.body, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.request() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.request() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPost(t *testing.T) {
	type fields struct {
		AutoRefresh bool
		ExpiresAt   time.Time
		HttpClient  HttpClient
	}
	type args struct {
		url    string
		body   map[string]any
		params map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Response
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			want: Response{
				body:   []byte{},
				status: 200,
			},
		},
		{
			name: "success with auto authenticate",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						if strings.Contains(req.URL.String(), "token") {
							return &http.Response{
								StatusCode: 200,
								Body:       io.NopCloser(bytes.NewReader([]byte(`{"token":"123"}`))),
							}, nil
						}

						return &http.Response{
							StatusCode: 200,
							Body:       http.NoBody,
						}, nil
					},
				},
				AutoRefresh: true,
			},
			want: Response{
				body:   []byte{},
				status: 200,
			},
		},
		{
			name: "fail to do request",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fail to auto authenticate",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
				AutoRefresh: true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				AutoRefresh: tt.fields.AutoRefresh,
				ExpiresAt:   tt.fields.ExpiresAt,
				HttpClient:  tt.fields.HttpClient,
			}
			got, err := client.post(context.Background(), tt.args.url, tt.args.body, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.post() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.post() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type fields struct {
		AutoRefresh bool
		ExpiresAt   time.Time
		HttpClient  HttpClient
	}
	type args struct {
		url    string
		params map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Response
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			want: Response{
				body:   []byte{},
				status: 200,
			},
		},
		{
			name: "success with auto authenticate",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						if strings.Contains(req.URL.String(), "token") {
							return &http.Response{
								StatusCode: 200,
								Body:       io.NopCloser(bytes.NewReader([]byte(`{"token":"123"}`))),
							}, nil
						}

						return &http.Response{
							StatusCode: 200,
							Body:       http.NoBody,
						}, nil
					},
				},
				AutoRefresh: true,
			},
			want: Response{
				body:   []byte{},
				status: 200,
			},
		},
		{
			name: "fail to do request",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fail to auto authenticate",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
				AutoRefresh: true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				AutoRefresh: tt.fields.AutoRefresh,
				ExpiresAt:   tt.fields.ExpiresAt,
				HttpClient:  tt.fields.HttpClient,
			}
			got, err := client.get(context.Background(), tt.args.url, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_UploadFile(t *testing.T) {
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		ctx  context.Context
		url  string
		body io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
		},
		{
			name: "invalid status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "fail to do request",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "fail to mount request",
			fields: fields{
				HttpClient: &ClientMock{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				HttpClient: tt.fields.HttpClient,
			}
			if err := client.uploadFile(tt.args.ctx, tt.args.url, tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("client.UploadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthenticate(t *testing.T) {
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		ctx          context.Context
		clientID     string
		clientSecret string
		expires      int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"token":"123"}`))),
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
		},
		{
			name: "fail to mount request",
			fields: fields{
				HttpClient: &ClientMock{},
			},
			wantErr: true,
		},
		{
			name: "fail to do request",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "invalid status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 403,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				HttpClient: tt.fields.HttpClient,
			}
			if err := client.Authenticate(tt.args.ctx, tt.args.clientID, tt.args.clientSecret, tt.args.expires); (err != nil) != tt.wantErr {
				t.Errorf("client.Authenticate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateSignedUrl(t *testing.T) {
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		service  string
		resource string
		metadata map[string]any
		params   map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    SignedUrlResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url","urls":{},"exp": 60000}`))),
						}, nil
					},
				},
			},
			want: SignedUrlResponse{
				Id:        "123",
				StatusURL: "url",
				Expires:   60000,
				URLs:      map[string]string{},
			},
		},
		{
			name: "invalid status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fail doing request",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				HttpClient: tt.fields.HttpClient,
			}
			got, err := client.GenerateSignedUrl(context.Background(), tt.args.service, tt.args.resource, tt.args.metadata, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.GenerateSignedUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.GenerateSignedUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUploadFileBase64(t *testing.T) {
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		url  string
		data string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				data: "123",
			},
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
		},
		{
			name: "fail doing request",
			args: args{
				data: "123",
			},
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				HttpClient: tt.fields.HttpClient,
			}
			if err := client.UploadFileBase64(context.Background(), tt.args.url, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("client.UploadFileBase64() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUploadFile(t *testing.T) {
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		url  string
		path string
	}
	f, _ := os.CreateTemp(".", "")
	defer os.Remove(f.Name())
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				path: f.Name(),
			},
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
		},
		{
			name: "fail doing request",
			args: args{
				path: f.Name(),
			},
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fail to open file",
			args: args{
				path: f.Name() + "1",
			},
			fields:  fields{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				HttpClient: tt.fields.HttpClient,
			}
			if err := client.UploadFile(context.Background(), tt.args.url, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("client.UploadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetBatchStatus(t *testing.T) {
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		batchID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    BatchStatusResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"batch_ksuid":"123","created_at":"2024-01-01","status":"done","service":"rg","jobs":[{"job_ksuid":"1234","created_at":"2024-01-01","status":"done","result_url":"url"}]}`))),
						}, nil
					},
				},
			},
			want: BatchStatusResponse{
				BatchID:   "123",
				CreatedAt: "2024-01-01",
				Service:   "rg",
				Status:    "done",
				Jobs: []BatchStatusJobs{
					{
						JobID:     "1234",
						CreatedAt: "2024-01-01",
						ResultURL: "url",
						Status:    "done",
					},
				},
			},
		},
		{
			name: "failed doing request",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("123")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 403,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				HttpClient: tt.fields.HttpClient,
			}
			got, err := client.GetBatchStatus(context.Background(), tt.args.batchID)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.GetBatchStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.GetBatchStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetJobResult(t *testing.T) {
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		batchID string
		jobID   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    JobResultResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"job_ksuid":"1234","created_at":"2024-01-01","status":"done","service":"rg"}`))),
						}, nil
					},
				},
			},
			want: JobResultResponse{
				JobID:     "1234",
				CreatedAt: "2024-01-01",
				Service:   "rg",
				Status:    "done",
			},
		},
		{
			name: "failed doing request",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("123")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 403,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				HttpClient: tt.fields.HttpClient,
			}
			got, err := client.GetJobResult(context.Background(), tt.args.batchID, tt.args.jobID)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.GetJobResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.GetJobResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetJobs(t *testing.T) {
	a := 0
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		start string
		end   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []JobResultResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						if a == 0 {
							a += 1
							return &http.Response{
								StatusCode: 200,
								Body:       io.NopCloser(bytes.NewReader([]byte(`{"jobs":[{"job_ksuid":"1234","created_at":"2024-01-01","status":"done","service":"rg"}], "nextPageToken": "1234"}`))),
							}, nil
						}
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"jobs":[{"job_ksuid":"12345","created_at":"2024-01-02","status":"done","service":"rg"}]}`))),
						}, nil
					},
				},
			},
			want: []JobResultResponse{
				{
					JobID:     "1234",
					CreatedAt: "2024-01-01",
					Service:   "rg",
					Status:    "done",
				},
				{
					JobID:     "12345",
					CreatedAt: "2024-01-02",
					Service:   "rg",
					Status:    "done",
				},
			},
		},
		{
			name: "failed doing request",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("123")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 403,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				HttpClient: tt.fields.HttpClient,
			}
			got, err := client.GetJobs(context.Background(), tt.args.start, tt.args.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.GetJobs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.GetJobs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendJobSingleStep(t *testing.T) {
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		service       string
		file          string
		facematchFile string
		extraFile     string
		metadata      map[string]any
		params        map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    CreatedResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			want: CreatedResponse{
				Id:        "123",
				StatusURL: "url/123",
			},
		},
		{
			name: "success with facematch and extra",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			args: args{
				params: map[string]string{
					"extra-document": "true",
					"facematch":      "true",
				},
			},
			want: CreatedResponse{
				Id:        "123",
				StatusURL: "url/123",
			},
		},
		{
			name: "failed doing request",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				HttpClient: tt.fields.HttpClient,
			}
			got, err := client.SendJobSingleStep(context.Background(), tt.args.service, tt.args.file, tt.args.facematchFile, tt.args.extraFile, tt.args.metadata, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.SendJobSingleStep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.SendJobSingleStep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendJobBase64(t *testing.T) {
	a := 0
	b := 0
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		service       string
		file          string
		facematchFile string
		extraFile     string
		metadata      map[string]any
		params        map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    CreatedResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			want: CreatedResponse{
				Id:        "123",
				StatusURL: "url/123",
			},
		},
		{
			name: "success with facematch and extra",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			args: args{
				params: map[string]string{
					"extra-document": "true",
					"facematch":      "true",
				},
			},
			want: CreatedResponse{
				Id:        "123",
				StatusURL: "url/123",
			},
		},
		{
			name: "failed doing request",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "failed to upload file",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						if req.Method == "PUT" {
							return nil, errors.New("error")
						}

						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "failed to upload facematch file",
			args: args{
				params: map[string]string{
					"extra-document": "true",
					"facematch":      "true",
				},
			},
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						if req.Method == "PUT" {
							a += 1
							if a == 2 {
								return nil, errors.New("error")
							}
						}

						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "failed to upload extra file",
			args: args{
				params: map[string]string{
					"extra-document": "true",
					"facematch":      "true",
				},
			},
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						if req.Method == "PUT" {
							b += 1
							if b == 3 {
								return nil, errors.New("error")
							}
						}

						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				HttpClient: tt.fields.HttpClient,
			}
			got, err := client.SendJobBase64(context.Background(), tt.args.service, tt.args.file, tt.args.facematchFile, tt.args.extraFile, tt.args.metadata, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.SendJobBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.SendJobBase64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendJob(t *testing.T) {
	a := 0
	b := 0
	f, _ := os.CreateTemp(".", "")
	defer os.Remove(f.Name())
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		service  string
		metadata map[string]any
		params   map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    CreatedResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			want: CreatedResponse{
				Id:        "123",
				StatusURL: "url/123",
			},
		},
		{
			name: "success with facematch and extra",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			args: args{
				params: map[string]string{
					"extra-document": "true",
					"facematch":      "true",
				},
			},
			want: CreatedResponse{
				Id:        "123",
				StatusURL: "url/123",
			},
		},
		{
			name: "failed doing request",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "failed to upload file",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						if req.Method == "PUT" {
							return nil, errors.New("error")
						}

						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "failed to upload facematch file",
			args: args{
				params: map[string]string{
					"extra-document": "true",
					"facematch":      "true",
				},
			},
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						if req.Method == "PUT" {
							a += 1
							if a == 2 {
								return nil, errors.New("error")
							}
						}

						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "failed to upload extra file",
			args: args{
				params: map[string]string{
					"extra-document": "true",
					"facematch":      "true",
				},
			},
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						if req.Method == "PUT" {
							b += 1
							if b == 3 {
								return nil, errors.New("error")
							}
						}

						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				HttpClient: tt.fields.HttpClient,
			}
			got, err := client.SendJob(context.Background(), tt.args.service, f.Name(), f.Name(), f.Name(), tt.args.metadata, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.SendJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.SendJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendBatchBase64(t *testing.T) {
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		service  string
		file     string
		metadata map[string]any
		params   map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    CreatedResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			want: CreatedResponse{
				Id:        "123",
				StatusURL: "url/123",
			},
		},
		{
			name: "failed doing request",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "failed to upload file",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						if req.Method == "PUT" {
							return nil, errors.New("error")
						}

						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				HttpClient: tt.fields.HttpClient,
			}
			got, err := client.SendBatchBase64(context.Background(), tt.args.service, tt.args.file, tt.args.metadata, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.SendBatchBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.SendBatchBase64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendBatch(t *testing.T) {
	f, _ := os.CreateTemp(".", "")
	defer os.Remove(f.Name())
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		service  string
		metadata []map[string]any
		params   map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    CreatedResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			want: CreatedResponse{
				Id:        "123",
				StatusURL: "url/123",
			},
		},
		{
			name: "failed doing request",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "failed to upload file",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						if req.Method == "PUT" {
							return nil, errors.New("error")
						}

						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status_url":"url/123"}`))),
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				HttpClient: tt.fields.HttpClient,
			}
			got, err := client.SendBatch(context.Background(), tt.args.service, f.Name(), tt.args.metadata, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.SendBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.SendBatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWaitForJobDone(t *testing.T) {
	type fields struct {
		Timeout    int
		Interval   int
		HttpClient HttpClient
	}
	type args struct {
		batchID string
		jobID   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    JobResultResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"job_ksuid":"123","status":"done"}`))),
						}, nil
					},
				},
			},
			want: JobResultResponse{
				JobID:  "123",
				Status: "done",
			},
		},
		{
			name: "invalid status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "timeout",
			fields: fields{
				Timeout:  1,
				Interval: 1,
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status":"processing"}`))),
						}, nil
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				Timeout:    tt.fields.Timeout,
				Interval:   tt.fields.Interval,
				HttpClient: tt.fields.HttpClient,
			}
			got, err := client.WaitForJobDone(context.Background(), tt.args.batchID, tt.args.jobID)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.WaitForJobDone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.WaitForJobDone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWaitForBatchDone(t *testing.T) {
	a := 0
	type fields struct {
		Timeout    int
		Interval   int
		HttpClient HttpClient
	}
	type args struct {
		batchID  string
		waitJobs bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    BatchStatusResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"batch_ksuid":"123","status":"done"}`))),
						}, nil
					},
				},
			},
			want: BatchStatusResponse{
				BatchID: "123",
				Status:  "done",
			},
		},
		{
			name: "success with wait jobs",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"batch_ksuid":"123","status":"done","jobs":[{"job_ksuid":"1234","status":"done"}]}`))),
						}, nil
					},
				},
			},
			args: args{
				waitJobs: true,
			},
			want: BatchStatusResponse{
				BatchID: "123",
				Status:  "done",
				Jobs: []BatchStatusJobs{
					{
						Status: "done",
						JobID:  "1234",
					},
				},
			},
		},
		{
			name: "failed to wait jobs",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						a += 1
						if a == 1 {
							return &http.Response{
								StatusCode: 200,
								Body:       io.NopCloser(bytes.NewReader([]byte(`{"batch_ksuid":"123","status":"done","jobs":[{"job_ksuid":"1234","status":"processing"}]}`))),
							}, nil
						}
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"job_ksuid":"1234","status":"processing"}`))),
						}, nil
					},
				},
				Timeout:  1,
				Interval: 1,
			},
			args: args{
				waitJobs: true,
			},
			wantErr: true,
		},
		{
			name: "invalid status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "timeout",
			fields: fields{
				Timeout:  1,
				Interval: 1,
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"123","status":"processing"}`))),
						}, nil
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				Timeout:    tt.fields.Timeout,
				Interval:   tt.fields.Interval,
				HttpClient: tt.fields.HttpClient,
			}
			got, err := client.WaitForBatchDone(context.Background(), tt.args.batchID, tt.args.waitJobs)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.WaitForBatchDone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.WaitForBatchDone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateAndWaitJob(t *testing.T) {
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		service           string
		filePath          string
		facematchFilePath string
		extraFilePath     string
		metadata          map[string]any
		params            map[string]string
	}
	f, _ := os.CreateTemp(".", "")
	defer os.Remove(f.Name())
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    JobResultResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"job_ksuid":"123","status":"done"}`))),
						}, nil
					},
				},
			},
			args: args{
				filePath: f.Name(),
			},
			want: JobResultResponse{
				JobID:  "123",
				Status: "done",
			},
		},
		{
			name: "invalid status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid wait status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						if req.Method == "PUT" {
							return &http.Response{
								StatusCode: 200,
								Body:       http.NoBody,
							}, nil
						}

						return &http.Response{
							StatusCode: 500,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				HttpClient: tt.fields.HttpClient,
			}
			got, err := client.CreateAndWaitJob(context.Background(), tt.args.service, tt.args.filePath, tt.args.facematchFilePath, tt.args.extraFilePath, tt.args.metadata, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.CreateAndWaitJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.CreateAndWaitJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateAndWaitBatch(t *testing.T) {
	type fields struct {
		HttpClient HttpClient
	}
	type args struct {
		service  string
		filePath string
		metadata []map[string]any
		params   map[string]string
		waitJobs bool
	}
	f, _ := os.CreateTemp(".", "")
	defer os.Remove(f.Name())
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    BatchStatusResponse
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(`{"batch_ksuid":"123","status":"done"}`))),
						}, nil
					},
				},
			},
			args: args{
				filePath: f.Name(),
			},
			want: BatchStatusResponse{
				BatchID: "123",
				Status:  "done",
			},
		},
		{
			name: "invalid status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid wait status code",
			fields: fields{
				HttpClient: &ClientMock{
					MockDo: func(req *http.Request) (*http.Response, error) {
						if req.Method == "PUT" {
							return &http.Response{
								StatusCode: 200,
								Body:       http.NoBody,
							}, nil
						}

						return &http.Response{
							StatusCode: 500,
							Body:       http.NoBody,
						}, nil
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				HttpClient: tt.fields.HttpClient,
			}
			got, err := client.CreateAndWaitBatch(context.Background(), tt.args.service, tt.args.filePath, tt.args.metadata, tt.args.params, tt.args.waitJobs)
			if (err != nil) != tt.wantErr {
				t.Errorf("client.CreateAndWaitBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.CreateAndWaitBatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
