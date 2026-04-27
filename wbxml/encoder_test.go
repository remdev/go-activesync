package wbxml

import (
	"bytes"
	"testing"
)

// SPEC: MS-ASWBXML/encoder.tag
// SPEC: MS-ASWBXML/encoder.end
func TestEncoder_StartEndTag_SamePage(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A}); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	// AirSync.Sync (page 0, identity 0x05) with content
	if err := enc.StartTag(PageAirSync, 0x05, false, true); err != nil {
		t.Fatalf("StartTag: %v", err)
	}
	if err := enc.EndTag(); err != nil {
		t.Fatalf("EndTag: %v", err)
	}
	want := []byte{0x03, 0x01, 0x6A, 0x00, 0x45, 0x01}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("bytes = % X, want % X", buf.Bytes(), want)
	}
}

// SPEC: MS-ASWBXML/encoder.switch-page
func TestEncoder_SwitchPage(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A}); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	if err := enc.StartTag(PageAirSync, 0x05, false, true); err != nil {
		t.Fatalf("StartTag AirSync: %v", err)
	}
	// AirSyncBase is page 17, so encoder must inject SWITCH_PAGE 0x00 0x11
	if err := enc.StartTag(PageAirSyncBase, 0x0A, false, true); err != nil {
		t.Fatalf("StartTag AirSyncBase: %v", err)
	}
	if err := enc.EndTag(); err != nil {
		t.Fatalf("EndTag inner: %v", err)
	}
	if err := enc.EndTag(); err != nil {
		t.Fatalf("EndTag outer: %v", err)
	}
	want := []byte{0x03, 0x01, 0x6A, 0x00, 0x45, 0x00, 0x11, 0x4A, 0x01, 0x01}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("bytes = % X, want % X", buf.Bytes(), want)
	}
}

// SPEC: MS-ASWBXML/encoder.string
func TestEncoder_StrI(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.StrI("ok"); err != nil {
		t.Fatalf("StrI: %v", err)
	}
	want := []byte{0x03, 'o', 'k', 0x00}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("bytes = % X, want % X", buf.Bytes(), want)
	}
}

// SPEC: MS-ASWBXML/encoder.opaque
func TestEncoder_Opaque(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Opaque([]byte{0xDE, 0xAD, 0xBE, 0xEF}); err != nil {
		t.Fatalf("Opaque: %v", err)
	}
	want := []byte{0xC3, 0x04, 0xDE, 0xAD, 0xBE, 0xEF}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("bytes = % X, want % X", buf.Bytes(), want)
	}
}

// SPEC: MS-ASWBXML/encoder.tag
func TestEncoder_RejectsUnknownPage(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.StartTag(0xFE, 0x05, false, true); err == nil {
		t.Fatalf("StartTag with unknown page: expected error, got nil")
	}
}
