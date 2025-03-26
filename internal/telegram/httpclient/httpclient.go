package httpclient

import (
	"errors"
	"fmt"
	"net/url"

	"resty.dev/v3"
)

type httpClient struct {
	restyClient *resty.Client
}

type successResponse struct {
	OK bool `json:"ok"`
}

type errorResponse struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}

func New(apiBotURL, token string) (HTTPDoer, error) {
	if apiBotURL == "" {
		return nil, errors.New("apiBotURL value is not provided")
	}

	if token == "" {
		return nil, errors.New("token value is not provided")
	}

	baseURLWithToken, err := url.JoinPath(apiBotURL, "/bot"+token)
	if err != nil {
		return nil, fmt.Errorf("invalid API URL: %w", err)
	}

	restyClient := resty.New().
		// SetDebug(true).
		SetBaseURL(baseURLWithToken)

	return httpClient{
		restyClient: restyClient,
	}, nil
}

func (c httpClient) SubmitMultipart(method, endpoint string, fields map[string]string, files []MultipartFile) error {
	request := c.restyClient.R()
	request.SetMultipartFormData(fields)

	for _, mFile := range files {
		request.SetFileReader(mFile.FieldName, mFile.FileName, mFile.FileReader)
	}

	return c.executeRequest(method, endpoint, request)
}

func (c httpClient) SubmitJSON(method, endpoint string, body any) error {
	request := c.restyClient.R()
	request.SetBody(body)

	return c.executeRequest(method, endpoint, request)
}

// executeRequest executes the request and handles the response.
func (c httpClient) executeRequest(method, endpoint string, request *resty.Request) error {
	successPayload := &successResponse{}
	errorPayload := &errorResponse{}

	resp, err := request.
		SetResult(successPayload).
		SetError(errorPayload).
		Execute(method, endpoint)
	if err != nil {
		return fmt.Errorf("transport error: %w", err)
	}

	if resp.IsSuccess() && successPayload.OK {
		return nil
	}

	if errorPayload.Description != "" {
		return fmt.Errorf("API error [%s] (HTTP %d): %d - %s",
			endpoint, resp.StatusCode(), errorPayload.ErrorCode, errorPayload.Description)
	}

	body := resp.String()
	if body == "" {
		body = "<empty response body>"
	}

	return fmt.Errorf("unexpected error (status=%d): %s", resp.StatusCode(), body)
}
