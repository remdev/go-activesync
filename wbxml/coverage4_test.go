package wbxml

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

// plainReader exposes only io.Reader, never io.ByteReader.
type plainReader struct{ r io.Reader }

func (p *plainReader) Read(b []byte) (int, error) { return p.r.Read(b) }

// SPEC: MS-ASWBXML/decoder.tag
func TestByteReader_FallbackToBufio(t *testing.T) {
	pr := &plainReader{r: bytes.NewReader([]byte{0x03, 0x01, 0x6A, 0x00})}
	var h Header
	if err := h.Read(pr); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if h.Version != 0x03 {
		t.Fatalf("Version = %x", h.Version)
	}
}

// stallReader returns a non-EOF error on the first Read so that
// io.ReadFull surfaces it during string-table reads.
type stallReader struct {
	r       io.Reader
	stalled bool
}

func (s *stallReader) Read(b []byte) (int, error) {
	if !s.stalled {
		s.stalled = true
		return s.r.Read(b)
	}
	return 0, errors.New("read stall")
}

// SPEC: OMA-WBXML-1.3/header.version
func TestHeader_StringTableReadError(t *testing.T) {
	src := bytes.NewReader([]byte{0x03, 0x01, 0x6A, 0x05, 'a', 'b'})
	if err := (&Header{}).Read(&stallReader{r: src}); err == nil {
		t.Fatal("expected string-table read error")
	}
}

// SPEC: MS-ASWBXML/encoder.switch-page
type firstWriteOK struct {
	allowed int
}

func (f *firstWriteOK) Write(p []byte) (int, error) {
	if f.allowed <= 0 {
		return 0, errors.New("blocked")
	}
	f.allowed--
	return len(p), nil
}

func TestHeader_WriteError(t *testing.T) {
	if err := (&Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A}).Write(&firstWriteOK{}); err == nil {
		t.Fatal("expected write failure")
	}
}

// SPEC: MS-ASWBXML/encoder.switch-page
func TestEncoder_StartTagSwitchError(t *testing.T) {
	enc := NewEncoder(&firstWriteOK{})
	if err := enc.StartTag(2 /* Email */, 0x05, false, false); err == nil {
		t.Fatal("expected SWITCH_PAGE write failure")
	}
}
