package httpclient

import (
	"context"
	"io"
)

type MultipartFile struct {
	FieldName  string
	FileName   string
	FileReader io.ReadCloser
}

// HTTPDoer is an interface to abstract away the underlying HTTP client logic.
type HTTPDoer interface {
	SubmitMultipart(ctx context.Context, method, endpoint string, fields map[string]string, files []MultipartFile) error
	SubmitJSON(ctx context.Context, method, endpoint string, body any) error
}
