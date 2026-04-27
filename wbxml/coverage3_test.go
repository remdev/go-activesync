package wbxml

import (
	"bufio"
	"bytes"
	"reflect"
	"strings"
	"testing"
)

// SPEC: MS-ASWBXML/marshal.roundtrip
type byteSlicePlain struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Class   []byte   `wbxml:"AirSync.Class"`
}

func TestMarshal_ByteSliceNonOpaque(t *testing.T) {
	in := byteSlicePlain{Class: []byte("Email")}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out byteSlicePlain
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if string(out.Class) != "Email" {
		t.Fatalf("Class = %q", out.Class)
	}
}

// SPEC: MS-ASWBXML/marshal.slice
type sliceFloatPtr struct {
	XMLName struct{}    `wbxml:"AirSync.Sync"`
	Vals    []*float64  `wbxml:"AirSync.SyncKey"`
}

func TestMarshal_SliceUnsupportedPtrElem(t *testing.T) {
	v := 1.0
	in := sliceFloatPtr{Vals: []*float64{&v}}
	if _, err := Marshal(&in); err == nil {
		t.Fatal("expected unsupported slice ptr elem error")
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
func TestMarshal_NilTopLevel(t *testing.T) {
	if _, err := Marshal(nil); err == nil {
		t.Fatal("expected error for nil")
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestDecodeField_UnexpectedToken(t *testing.T) {
	// Construct a doc whose Sync content begins with a STR_I instead of a tag.
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A})
	enc.StartTag(0, 0x05 /* Sync */, false, true)
	enc.StrI("loose")
	enc.EndTag()

	if err := Unmarshal(buf.Bytes(), &wrappedTag{}); err == nil {
		t.Fatal("expected unexpected-token error inside struct")
	}
}

// SPEC: MS-ASWBXML/decoder.tag
func TestByteReader_BufferedReader(t *testing.T) {
	br := bufio.NewReader(strings.NewReader("ab"))
	got := byteReader(br)
	if got != br {
		t.Fatal("expected pass-through")
	}
}

// SPEC: MS-ASWBXML/codepage.invariants
func TestPageByID_Unknown(t *testing.T) {
	if _, ok := PageByID(99); ok {
		t.Fatal("expected unknown page miss")
	}
}

// SPEC: MS-ASWBXML/codepage.invariants
func TestPageByName_Unknown(t *testing.T) {
	if _, ok := PageByName("Bogus"); ok {
		t.Fatal("expected unknown page miss")
	}
}

// SPEC: MS-ASWBXML/marshal.omitempty
func TestIsZero_MapAndPointer(t *testing.T) {
	m := map[string]int(nil)
	if !isZero(reflect.ValueOf(m)) {
		t.Fatal("nil map not zero")
	}
	mp := map[string]int{"a": 1}
	if isZero(reflect.ValueOf(mp)) {
		t.Fatal("populated map zero")
	}
	var ptr *int
	if !isZero(reflect.ValueOf(ptr)) {
		t.Fatal("nil pointer not zero")
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
func TestParseTag_EmptyOption(t *testing.T) {
	spec, err := parseTag("AirSync.Sync,,omitempty")
	if err != nil {
		t.Fatalf("parseTag: %v", err)
	}
	if !spec.omitempty {
		t.Fatal("omitempty not set")
	}
}
