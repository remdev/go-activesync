package wbxml

import (
	"bytes"
	"reflect"
	"testing"
)

// SPEC: MS-ASWBXML/marshal.omitempty
type ptrOmit struct {
	XMLName struct{}      `wbxml:"AirSync.Sync"`
	Inner   *innerScalar  `wbxml:"AirSync.Collection,omitempty"`
}

func TestMarshal_NilPointerWithOmitempty(t *testing.T) {
	in := ptrOmit{}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out ptrOmit
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Inner != nil {
		t.Fatalf("Inner should be nil")
	}
}

// SPEC: MS-ASWBXML/marshal.slice
type sliceBoolHolder struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Bs      []bool   `wbxml:"AirSync.GetChanges"`
}

func TestMarshal_SliceOfBoolTrue(t *testing.T) {
	in := sliceBoolHolder{Bs: []bool{true, false, true}}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out sliceBoolHolder
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in.Bs, out.Bs) {
		t.Fatalf("Bs = %v, want %v", out.Bs, in.Bs)
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
type byteSliceHolder struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Class   []byte   `wbxml:"AirSync.Class,opaque"`
}

func TestUnmarshal_OpaqueByteSliceWithoutContent(t *testing.T) {
	// Build doc: Sync (content) → Class (no content) → End.
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A})
	enc.StartTag(0, 0x05 /* Sync */, false, true)
	enc.StartTag(0, 0x10 /* Class */, false, false)
	enc.EndTag()
	var out byteSliceHolder
	if err := Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Class != nil {
		t.Fatalf("Class = %v", out.Class)
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
type sliceIntHolder struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Vals    []int32  `wbxml:"AirSync.WindowSize"`
}

func TestUnmarshal_SliceOfScalarWithoutContent(t *testing.T) {
	// Build doc: Sync (content) → WindowSize (no content) → end.
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A})
	enc.StartTag(0, 0x05 /* Sync */, false, true)
	enc.StartTag(0, 0x15 /* WindowSize */, false, false)
	enc.EndTag()

	var out sliceIntHolder
	if err := Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.Vals) != 1 {
		t.Fatalf("Vals = %v", out.Vals)
	}
}

// SPEC: MS-ASWBXML/marshal.slice
type sliceUnsupportedDecode struct {
	XMLName struct{}    `wbxml:"AirSync.Sync"`
	Vals    []float64   `wbxml:"AirSync.SyncKey"`
}

func TestUnmarshal_SliceUnsupportedKind(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A})
	enc.StartTag(0, 0x05 /* Sync */, false, true)
	enc.StartTag(0, 0x0B /* SyncKey */, false, true)
	enc.StrI("x")
	enc.EndTag()
	enc.EndTag()
	if err := Unmarshal(buf.Bytes(), &sliceUnsupportedDecode{}); err == nil {
		t.Fatal("expected unsupported slice elem kind error")
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
type unsupportedDecode struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Val     float64  `wbxml:"AirSync.SyncKey"`
}

func TestUnmarshal_UnsupportedFieldKind(t *testing.T) {
	// Hand-craft a doc that has Sync→SyncKey="x"→end. Decoding into float64
	// must produce an error from decodeField's default branch.
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteHeader(Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A})
	enc.StartTag(0, 0x05 /* Sync */, false, true)
	enc.StartTag(0, 0x0B /* SyncKey */, false, true)
	enc.StrI("1.5")
	enc.EndTag()
	enc.EndTag()

	if err := Unmarshal(buf.Bytes(), &unsupportedDecode{}); err == nil {
		t.Fatal("expected unsupported field-kind error")
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
type noXMLName struct {
	K string `wbxml:"AirSync.SyncKey"`
}

func TestMarshal_NoXMLName(t *testing.T) {
	if _, err := Marshal(&noXMLName{}); err == nil {
		t.Fatal("expected missing XMLName error")
	}
}

func TestUnmarshal_NoXMLName(t *testing.T) {
	if err := Unmarshal([]byte{0x03, 0x01, 0x6A, 0x00, 0x05, 0x01}, &noXMLName{}); err == nil {
		t.Fatal("expected missing XMLName error")
	}
}

// SPEC: MS-ASWBXML/marshal.slice
type sliceUintHolder struct {
	XMLName struct{}  `wbxml:"AirSync.Sync"`
	Vals    []uint32  `wbxml:"AirSync.MIMETruncation"`
}

func TestMarshal_SliceOfUint(t *testing.T) {
	in := sliceUintHolder{Vals: []uint32{1, 2, 3}}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out sliceUintHolder
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in.Vals, out.Vals) {
		t.Fatalf("Vals = %v", out.Vals)
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
func TestMarshal_OmitEmptyZeroScalars(t *testing.T) {
	type allZero struct {
		XMLName struct{} `wbxml:"AirSync.Sync"`
		S       string   `wbxml:"AirSync.SyncKey,omitempty"`
		B       bool     `wbxml:"AirSync.GetChanges,omitempty"`
		I       int32    `wbxml:"AirSync.WindowSize,omitempty"`
		U       uint32   `wbxml:"AirSync.MIMETruncation,omitempty"`
	}
	in := allZero{}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out allZero
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round-trip mismatch:\n got %+v\nwant %+v", out, in)
	}
}
