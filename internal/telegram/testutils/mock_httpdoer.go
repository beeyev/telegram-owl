package testutils

import "github.com/beeyev/telegram-owl/internal/telegram/httpclient"

// MockHTTPDoer.
type MockHTTPDoer struct {
	SubmitMultipartResult []submitMultipartPayload
	SubmitJSONResult      []submitJSONPayload
}

type submitMultipartPayload struct {
	Method   string
	Endpoint string
	Fields   map[string]string
	Files    []httpclient.MultipartFile
}

type submitJSONPayload struct {
	Method   string
	Endpoint string
	Body     any
}

func (c *MockHTTPDoer) SubmitMultipart(
	method,
	endpoint string,
	fields map[string]string,
	files []httpclient.MultipartFile,
) error {
	c.SubmitMultipartResult = append(c.SubmitMultipartResult, submitMultipartPayload{
		Method:   method,
		Endpoint: endpoint,
		Fields:   fields,
		Files:    files,
	})

	return nil
}

func (c *MockHTTPDoer) SubmitJSON(method, endpoint string, body any) error {
	c.SubmitJSONResult = append(c.SubmitJSONResult, submitJSONPayload{
		Method:   method,
		Endpoint: endpoint,
		Body:     body,
	})

	return nil
}

func NewMockHTTPDoer() *MockHTTPDoer {
	return &MockHTTPDoer{}
}
