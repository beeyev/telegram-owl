package httpclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"resty.dev/v3"
)

const defaultRequestTimeout = 30 * time.Second

type httpClient struct {
	restyClient *resty.Client
}

type successResponse struct {
	OK bool `json:"ok"`
}

type errorResponse struct {
	OK          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}

func New(apiBotURL, token, proxyURL string) (HTTPDoer, error) {
	if apiBotURL == "" {
		return nil, errors.New("api bot url value is not provided")
	}

	if token == "" {
		return nil, errors.New("token value is not provided")
	}

	baseURLWithToken, err := url.JoinPath(apiBotURL, "/bot"+token)
	if err != nil {
		return nil, fmt.Errorf("invalid api url: %w", err)
	}

	restyClient := resty.New().
		// SetDebug(true).
		SetBaseURL(baseURLWithToken).
		SetTimeout(defaultRequestTimeout)

	if proxyURL != "" {
		if _, err = url.ParseRequestURI(proxyURL); err != nil {
			return nil, fmt.Errorf("invalid proxy url: %w", err)
		}

		restyClient.SetProxy(proxyURL)
	}

	return httpClient{
		restyClient: restyClient,
	}, nil
}

func (c httpClient) SubmitMultipart(
	ctx context.Context,
	method,
	endpoint string,
	fields map[string]string,
	files []MultipartFile,
) error {
	request := c.restyClient.R()
	request.SetMultipartFormData(fields)

	for _, mFile := range files {
		request.SetFileReader(mFile.FieldName, mFile.FileName, hideCloser(mFile.FileReader))
	}

	return c.executeRequest(ctx, method, endpoint, request)
}

// hideCloser keeps reader ownership with the caller while preserving seek support.
func hideCloser(reader io.Reader) io.Reader {
	if readSeeker, ok := reader.(io.ReadSeeker); ok {
		return struct{ io.ReadSeeker }{ReadSeeker: readSeeker}
	}

	return struct{ io.Reader }{Reader: reader}
}

func (c httpClient) SubmitJSON(ctx context.Context, method, endpoint string, body any) error {
	request := c.restyClient.R()
	request.SetBody(body)

	return c.executeRequest(ctx, method, endpoint, request)
}

// executeRequest executes the request and handles the response.
func (c httpClient) executeRequest(ctx context.Context, method, endpoint string, request *resty.Request) error {
	if ctx == nil {
		return errors.New("context is nil")
	}

	successPayload := &successResponse{}
	errorPayload := &errorResponse{}

	resp, err := request.
		SetContext(ctx).
		SetResult(successPayload).
		SetResultError(errorPayload).
		Execute(method, endpoint)
	if err != nil {
		if urlErr, ok := errors.AsType[*url.Error](err); ok {
			return fmt.Errorf("telegram api request failed: %w", urlErr.Err)
		}

		return fmt.Errorf("telegram api request failed: %w", err)
	}

	if resp.IsStatusSuccess() && successPayload.OK {
		return nil
	}

	if errorPayload.Description != "" {
		return fmt.Errorf("telegram api error [%s] (http %d): %d - %s",
			endpoint, resp.StatusCode(), errorPayload.ErrorCode, errorPayload.Description)
	}

	body := resp.String()
	if body == "" {
		body = "<empty response body>"
	}

	return fmt.Errorf("unexpected error (status=%d): %s", resp.StatusCode(), body)
}
