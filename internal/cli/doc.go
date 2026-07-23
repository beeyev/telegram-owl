// Package cli defines the telegram-owl command surface and coordinates input,
// attachment loading, Telegram requests, and user-facing output.
//
// Keep transport details out of this package. The CLI decides what to send and
// owns opened attachments; the telegram packages decide how to encode requests.
package cli
