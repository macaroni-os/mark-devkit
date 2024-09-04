/*
Copyright © 2021-2023 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package utils

import "bytes"

type NopCloseWriter struct {
	*bytes.Buffer
}

func NewNopCloseWriter(buf *bytes.Buffer) *NopCloseWriter {
	return &NopCloseWriter{Buffer: buf}
}

func (ncw *NopCloseWriter) Close() error { return nil }
