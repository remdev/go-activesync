package wbxml

import (
	"bufio"
	"io"
)

// byteWriter adapts an io.Writer to the io.ByteWriter interface so that
// callers can hand any writer to mb_u_int32 helpers.
func byteWriter(w io.Writer) io.ByteWriter {
	if bw, ok := w.(io.ByteWriter); ok {
		return bw
	}
	return bufio.NewWriter(w)
}

// byteReader adapts an io.Reader to the io.ByteReader interface.
func byteReader(r io.Reader) io.ByteReader {
	if br, ok := r.(io.ByteReader); ok {
		return br
	}
	return bufio.NewReader(r)
}
