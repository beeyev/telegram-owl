// Package attachment classifies, validates, and opens files for Telegram media
// requests.
//
// A successful Loader call transfers ownership of every open file to the
// caller. A failed call closes everything opened so far. Call Attachments.Close
// exactly once after the synchronous send completes.
package attachment
