// Package httpclient adapts Telegram request payloads to Resty.
//
// Multipart readers are borrowed for the duration of SubmitMultipart. Their
// caller retains ownership and remains responsible for closing them.
package httpclient
