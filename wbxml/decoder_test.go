package wbxml

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

// SPEC: MS-ASWBXML/decoder.tag
// SPEC: MS-ASWBXML/decoder.switch-page
func TestDecoder_TagAndSwitchPage(t *testing.T) {
	// Header + <Sync> <Body in page 17> </Body> </Sync>
	in := []byte{0x03, 0x01, 0x6A, 0x00, 0x45, 0x00, 0x11, 0x4A, 0x01, 0x01}
	dec := NewDecoder(bytes.NewReader(in))
	h, err := dec.ReadHeader()
	if err != nil {
		t.Fatalf("ReadHeader: %v", err)
	}
	if h.Version != 0x03 || h.PublicID != 0x01 || h.Charset != 0x6A {
		t.Fatalf("Header = %+v", h)
	}

	tok, err := dec.NextToken()
	if err != nil {
		t.Fatalf("NextToken 1: %v", err)
	}
	if tok.Kind != KindTag || tok.Page != PageAirSync || tok.Tag != 0x05 || !tok.HasContent || tok.HasAttrs {
		t.Fatalf("token 1 = %+v", tok)
	}
	tok, err = dec.NextToken()
	if err != nil {
		t.Fatalf("NextToken 2: %v", err)
	}
	if tok.Kind != KindTag || tok.Page != PageAirSyncBase || tok.Tag != 0x0A {
		t.Fatalf("token 2 = %+v (want page=AirSyncBase tag=0x0A)", tok)
	}

	tok, err = dec.NextToken()
	if err != nil {
		t.Fatalf("NextToken 3: %v", err)
	}
	if tok.Kind != KindEnd {
		t.Fatalf("token 3 = %+v (want End)", tok)
	}
	tok, err = dec.NextToken()
	if err != nil {
		t.Fatalf("NextToken 4: %v", err)
	}
	if tok.Kind != KindEnd {
		t.Fatalf("token 4 = %+v (want End)", tok)
	}
	if _, err := dec.NextToken(); !errors.Is(err, io.EOF) {
		t.Fatalf("NextToken 5: want io.EOF, got %v", err)
	}
}

// SPEC: MS-ASWBXML/decoder.string
func TestDecoder_StrI(t *testing.T) {
	in := []byte{0x03, 'o', 'k', 0x00}
	dec := NewDecoder(bytes.NewReader(in))
	tok, err := dec.NextToken()
	if err != nil {
		t.Fatalf("NextToken: %v", err)
	}
	if tok.Kind != KindString || tok.String != "ok" {
		t.Fatalf("token = %+v", tok)
	}
}

// SPEC: MS-ASWBXML/decoder.opaque
func TestDecoder_Opaque(t *testing.T) {
	in := []byte{0xC3, 0x04, 0xDE, 0xAD, 0xBE, 0xEF}
	dec := NewDecoder(bytes.NewReader(in))
	tok, err := dec.NextToken()
	if err != nil {
		t.Fatalf("NextToken: %v", err)
	}
	if tok.Kind != KindOpaque {
		t.Fatalf("token kind = %v want Opaque", tok.Kind)
	}
	if !bytes.Equal(tok.Bytes, []byte{0xDE, 0xAD, 0xBE, 0xEF}) {
		t.Fatalf("bytes = % X", tok.Bytes)
	}
}
