# UltraOCR SDK Go

UltraOCR SDK for Golang.

[UltraOCR](https://ultraocr.com.br/) is a platform that assists in the document analysis process with AI.

For more details about the system, types of documents, routes and params, access our [documentation](https://docs.nuveo.ai/ocr/v2/).

## Instalation

First of all, you must install this package with:

```
go get github.com/nuveo/ultraocr-sdk-go
```

Then you must import the UltraOCR SDK in your code , with:

```go
import (
    github.com/nuveo/ultraocr-sdk-go/ultraocr
)
```

## Step by step

### First step - Client Creation and Authentication

With the UltraOCR SDK installed and imported, the first step is create the Client and authenticate, you have two ways to do it.

The first one, you can do it authenticating when you want:

```go
client := ultraocr.NewClient()
client.Authenticate(context.Background(), "YOUR_CLIENT_ID", "YOUR_CLIENT_SECRET", 60)
```

The third argument is `expires`, a number between `1` and `1440`, the Token time expiration in minutes.

Another way is setting the client to auto refresh. As example:

```go
client := ultraocr.NewClient()
client.SetAutoRefresh("YOUR_CLIENT_ID", "YOUR_CLIENT_SECRET", 60)
```

The Client have following customizations:

* `SetAutoRefresh(string, string, int)`: Set auto authentication as showed above.
* `SetBaseURL(string)`: Change the base url to send documents (Default UltraOCR url).
* `SetAuthBaseURL(string)`: Change the base url to authenticate (Default UltraOCR url).
* `SetTimeout(int)`: Change the pooling timeout in seconds (Default 30).
* `SetInterval(int)`: Change the pooling interval in seconds (Default 1).
* `SetHttpClient(HttpClient)`: Change the http client to requests (Default http.DefaultClient).

### Second step - Send Documents

With everything set up, you can send documents:

```go
client.SendJob(CONTEXT, "SERVICE", "FILE_PATH", "", "", METADATA, PARAMS) // Simple job
client.SendBatch(CONTEXT, "SERVICE", "FILE_PATH", METADATA, PARAMS) // Simple batch
client.SendJobBase64(CONTEXT, "SERVICE", "BASE64_DATA", "", "", METADATA, PARAMS) // Job in base64
client.SendBatchBase64(CONTEXT, "SERVICE", "BASE64_DATA", METADATA, PARAMS) // Batch in base64
client.SendJobSingleStep(CONTEXT, "SERVICE", "BASE64_DATA", "", "", METADATA, PARAMS) // Job in base64, faster, but with limits

```

Send batch response example:

```go
CreatedResponse{
    Id: "0ujsszwN8NRY24YaXiTIE2VWDTS",
	StatusURL: "https://ultraocr.apis.nuveo.ai/v2/ocr/batch/status/0ujsszwN8NRY24YaXiTIE2VWDTS",
}
```

Send job response example:

```go
CreatedResponse{
    Id: "0ujsszwN8NRY24YaXiTIE2VWDTS",
	StatusURL: "https://ultraocr.apis.nuveo.ai/v2/ocr/job/result/0ujsszwN8NRY24YaXiTIE2VWDTS",
}
```

For jobs, to send facematch file you must provide the files and request on query params.

Examples using CNH service and sending facematch and extra files:

```go
params := map[string]string{
    "extra-document": "true",
    "facematch": "true"
}

client.SendJob(CONTEXT, "SERVICE", "FILE_PATH", "FACEMATCH_FILE_PATH", "EXTRA_FILE_PATH", METADATA, params)
client.SendJobBase64(CONTEXT, "SERVICE", "BASE64_DATA", "FACEMATCH_BASE64_DATA", "EXTRA_BASE64_DATA", METADATA, params)
client.SendJobSingleStep(CONTEXT, "SERVICE", "BASE64_DATA", "FACEMATCH_BASE64_DATA", "EXTRA_BASE64_DATA", METADATA, params)
```

Alternatively, you can request the signed url directly, without any utility, but you will must to upload the document manually. Example:

```go
res, err := client.GenerateSignedUrl(CONTEXT, "SERVICE", "job", METADATA, PARAMS) // Request job
urls := response.URLs
url = urls["document"]

// Use utility to upload
err = client.UploadFile(ctx, url, "FILE_PATH")

res, err = client.GenerateSignedUrl(CONTEXT, "SERVICE", "batch", METADATA, PARAMS) // Request batch
urls = response.URLs
url = urls["document"]

// Manual upload
f, err := os.ReadFile("FILE_PATH")

req, err := http.NewRequestWithContext(CONTEXT, http.MethodPut, url, bytes.NewBuffer(f))
res, err = httpClient.Do(req)
```

Example of response from `GenerateSignedUrl` with facematch and extra files:

```go
SignedUrlResponse{
	Expires: 60000,
	Id: "0ujsszwN8NRY24YaXiTIE2VWDTS",
	StatusURL: "https://ultraocr.apis.nuveo.ai/v2/ocr/batch/status/0ujsszwN8NRY24YaXiTIE2VWDTS",
	URLs: map[string]string{
        "document": "https://presignedurldemo.s3.eu-west-2.amazonaws.com/image.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAJJWZ7B6WCRGMKFGQ%2F20180210%2Feu-west-2%2Fs3%2Faws4_request&X-Amz-Date=20180210T171315Z&X-Amz-Expires=1800&X-Amz-Signature=12b74b0788aa036bc7c3d03b3f20c61f1f91cc9ad8873e3314255dc479a25351&X-Amz-SignedHeaders=host",
        "selfie": "https://presignedurldemo.s3.eu-west-2.amazonaws.com/image.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAJJWZ7B6WCRGMKFGQ%2F20180210%2Feu-west-2%2Fs3%2Faws4_request&X-Amz-Date=20180210T171315Z&X-Amz-Expires=1800&X-Amz-Signature=12b74b0788aa036bc7c3d03b3f20c61f1f91cc9ad8873e3314255dc479a25351&X-Amz-SignedHeaders=host",
        "extra_document": "https://presignedurldemo.s3.eu-west-2.amazonaws.com/image.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAJJWZ7B6WCRGMKFGQ%2F20180210%2Feu-west-2%2Fs3%2Faws4_request&X-Amz-Date=20180210T171315Z&X-Amz-Expires=1800&X-Amz-Signature=12b74b0788aa036bc7c3d03b3f20c61f1f91cc9ad8873e3314255dc479a25351&X-Amz-SignedHeaders=host"
    },
}
```

### Third step - Get Result

With the job or batch id, you can get the job result or batch status with:

```go
client.GetBatchStatus(CONTEXT, "BATCH_ID") // Batches
client.GetJobResult(CONTEXT, "JOB_ID", "JOB_ID") // Simple jobs
client.GetJobResult(CONTEXT, "BATCH_ID", "JOB_ID") // Jobs belonging to batches
```

Alternatively, you can use a utily `WaitForJobDone` or `WaitForBatchDone`:

```go
client.WaitForBatchDone(CONTEXT, "BATCH_ID", true) // Batches, ends when the batch and all it jobs are finished
client.WaitForBatchDone(CONTEXT, "BATCH_ID", false) // Batches, ends when the batch is finished
client.WaitForJobDone(CONTEXT, "JOB_ID", "JOB_ID") // Simple jobs
client.WaitForJobDone(CONTEXT, "BATCH_ID", "JOB_ID") // Jobs belonging to batches
```

Batch status example:

```go
BatchStatusResponse{
	BatchID: "2AwrSd7bxEMbPrQ5jZHGDzQ4qL3",
	CreatedAt: "2022-06-22T20:58:09Z",
	Service: "cnh",
	Status: "done",
	Jobs: []BatchStatusJobs{
        {
            JobID: "0ujsszwN8NRY24YaXiTIE2VWDTS",
            CreatedAt: "2022-06-22T20:58:09Z",
            ResultURL: "https://ultraocr.apis.nuveo.ai/v2/ocr/job/result/2AwrSd7bxEMbPrQ5jZHGDzQ4qL3/0ujsszwN8NRY24YaXiTIE2VWDTS",
            Status: "processing",
        },
    },
}
```

Job result example:

```go
JobResultResponse{
	JobID: "2AwrSd7bxEMbPrQ5jZHGDzQ4qL3",
	CreatedAt: "2022-06-22T20:58:09Z",
	Service: "idtypification",
	Status: "done",
    Result: Result{
        Time: "7.45",
        Document: []map[string]any{
            {
                "Page": 1,
                "Data": map[string]any{
                    "DocumentType": map[string]any{
                        "conf": 99,
                        "value": "CNH"
                    }
                }
            },
        }
    },
}
```

### Simplified way

You can do all steps in a simplified way, with `CreateAndWaitJob` or `CreateAndWaitBatch` utilities:

```go
client := ultraocr.NewClient()
client.SetAutoRefresh("YOUR_CLIENT_ID", "YOUR_CLIENT_SECRET", 60)
client.CreateAndWaitJob(CONTEXT, "SERVICE", "FILE_PATH", "", "", METADATA, PARAMS)
```

Or:

```go
client := ultraocr.NewClient()
client.SetAutoRefresh("YOUR_CLIENT_ID", "YOUR_CLIENT_SECRET", 60)
client.CreateAndWaitBatch(CONTEXT, "SERVICE", "FILE_PATH", METADATA, PARAMS, false)
```

The `CreateAndWaitJob` has the `SendJob` arguments and `GetJobResult` response, while the `CreateAndWaitBatch` has the `SendBatch` arguments with the additional `waitJobs` in the end and `GetBatchStatus` response. 

### Get many results

You can get all jobs in a given interval by calling `GetJobs` utility:

```go
client.GetJobs(CONTEXT, "START_DATE", "END_DATE") // Dates in 2006-01-02 format (YYYY-MM-DD)
```

Results:

```go
[]JobResultResponse{
    {
        JobID: "2AwrSd7bxEMbPrQ5jZHGDzQ4qL3",
        CreatedAt: "2022-06-22T20:58:09Z",
        Service: "idtypification",
        Status: "done",
        Result: result{
            Time: "7.45",
            Document: []map[string]any{
                {
                    "Page": 1,
                    "Data": map[string]any{
                        "DocumentType": map[string]any{
                            "conf": 99,
                            "value": "CNH"
                        }
                    }
                },
            }
        },
    },
    {
        JobID: "2AwrSd7bxEMbPrQ5jZHGDzQ4qL4",
        CreatedAt: "2022-06-22T20:59:09Z",
        Service: "cnh",
        Status: "done",
        Result: result{
            Time: "8.45",
            Document: []map[string]any{
                {
                    "Page": 1,
                    "Data": map[string]any{
                        "DocumentType": map[string]any{
                            "conf": 99,
                            "value": "CNH"
                        }
                    }
                },
            }
        },
    },
    {
        JobID: "2AwrSd7bxEMbPrQ5jZHGDzQ4qL5",
        CreatedAt: "2022-06-22T20:59:39Z",
        Service: "cnh",
        Status: "processing",
    },
}
```