package httpclient

import (
	"context"
	"io"
)

// MultipartFile describes one borrowed reader in a multipart request. The
// HTTPDoer implementation must not close FileReader.
type MultipartFile struct {
	FieldName  string
	FileName   string
	FileReader io.Reader
}

// HTTPDoer is the transport boundary used by Telegram method packages.
type HTTPDoer interface {
	// SubmitMultipart borrows files until the call returns and never closes them.
	SubmitMultipart(ctx context.Context, method, endpoint string, fields map[string]string, files []MultipartFile) error
	// SubmitJSON encodes body as JSON and submits it to endpoint.
	SubmitJSON(ctx context.Context, method, endpoint string, body any) error
}
