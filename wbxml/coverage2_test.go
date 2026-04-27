package wbxml

import (
	"bytes"
	"reflect"
	"testing"
)

// SPEC: MS-ASWBXML/marshal.slice
type slicePtrNil struct {
	XMLName struct{}        `wbxml:"AirSync.Sync"`
	Items   []*innerScalar  `wbxml:"AirSync.Collection"`
}

type innerScalar struct {
	K string `wbxml:"AirSync.SyncKey"`
}

// SPEC: MS-ASWBXML/marshal.slice
func TestMarshal_SlicePtrWithNilEntries(t *testing.T) {
	in := slicePtrNil{Items: []*innerScalar{nil, {K: "x"}}}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out slicePtrNil
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.Items) != 1 || out.Items[0].K != "x" {
		t.Fatalf("Items = %+v", out.Items)
	}
}

// SPEC: MS-ASWBXML/marshal.slice
type sliceUnsupported struct {
	XMLName struct{}    `wbxml:"AirSync.Sync"`
	Vals    []float64   `wbxml:"AirSync.SyncKey"`
}

func TestMarshal_SliceOfUnsupportedKind(t *testing.T) {
	in := sliceUnsupported{Vals: []float64{1.0}}
	if _, err := Marshal(&in); err == nil {
		t.Fatal("expected unsupported slice elem error")
	}
}

// SPEC: MS-ASWBXML/marshal.tag-resolution
type unsupportedField struct {
	XMLName struct{} `wbxml:"AirSync.Sync"`
	Val     float64  `wbxml:"AirSync.SyncKey"`
}

func TestMarshal_UnsupportedKind(t *testing.T) {
	in := unsupportedField{Val: 1.5}
	if _, err := Marshal(&in); err == nil {
		t.Fatal("expected unsupported field error")
	}
}

// SPEC: MS-ASWBXML/encoder.switch-page
type writerErr struct{}

func (writerErr) Write(p []byte) (int, error) { return 0, errFailWrite }

var errFailWrite = &writeError{"forced"}

type writeError struct{ msg string }

func (e *writeError) Error() string { return e.msg }

func TestEncoder_StrIWriteError(t *testing.T) {
	enc := NewEncoder(writerErr{})
	if err := enc.StrI("x"); err == nil {
		t.Fatal("expected write error")
	}
}

func TestEncoder_OpaqueWriteError(t *testing.T) {
	enc := NewEncoder(writerErr{})
	if err := enc.Opaque([]byte{1}); err == nil {
		t.Fatal("expected write error")
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
type allOmit struct {
	XMLName struct{}  `wbxml:"AirSync.Sync"`
	Bytes   []byte    `wbxml:"AirSync.Class,omitempty"`
}

func TestMarshal_OmitEmptyBytes(t *testing.T) {
	in := allOmit{}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got allOmit
	if err := Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in, got) {
		t.Fatalf("round-trip mismatch")
	}
}

// SPEC: MS-ASWBXML/marshal.slice
type sliceStructInner struct {
	K string `wbxml:"AirSync.SyncKey"`
}
type sliceStructDirect struct {
	XMLName struct{}           `wbxml:"AirSync.Sync"`
	Items   []sliceStructInner `wbxml:"AirSync.Collection"`
}

func TestMarshal_SliceOfStructDirect(t *testing.T) {
	in := sliceStructDirect{Items: []sliceStructInner{{K: "a"}, {K: "b"}}}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out sliceStructDirect
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.Items) != 2 || out.Items[0].K != "a" || out.Items[1].K != "b" {
		t.Fatalf("Items = %+v", out.Items)
	}
}

// SPEC: MS-ASWBXML/marshal.roundtrip
type ptrStructHolder struct {
	XMLName struct{}          `wbxml:"AirSync.Sync"`
	Inner   *sliceStructInner `wbxml:"AirSync.Collection"`
}

func TestMarshal_PointerStructFilled(t *testing.T) {
	in := ptrStructHolder{Inner: &sliceStructInner{K: "p"}}
	data, err := Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out ptrStructHolder
	if err := Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Inner == nil || out.Inner.K != "p" {
		t.Fatalf("Inner = %+v", out.Inner)
	}
}

// SPEC: OMA-WBXML-1.3/header.version
func TestHeader_StringTableLargerStrings(t *testing.T) {
	src := Header{Version: 0x03, PublicID: 0x01, Charset: 0x6A, StringTable: bytes.Repeat([]byte{'a'}, 200)}
	src.StringTable = append(src.StringTable, 0)
	var buf bytes.Buffer
	if err := src.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	var got Header
	if err := got.Read(&buf); err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !bytes.Equal(got.StringTable, src.StringTable) {
		t.Fatal("StringTable mismatch")
	}
}

// SPEC: MS-ASWBXML/codepage.invariants
func TestCodepages_TagLookup(t *testing.T) {
	if name, ok := TagByToken(0, 0x05); !ok || name != "Sync" {
		t.Fatalf("TagByToken(0,0x05) = %q,%v", name, ok)
	}
	if _, ok := TagByToken(99, 0x05); ok {
		t.Fatal("expected unknown page miss")
	}
	if tok, ok := TokenByTag(0, "Sync"); !ok || tok != 0x05 {
		t.Fatalf("TokenByTag(0,Sync) = %02X,%v", tok, ok)
	}
	if _, ok := TokenByTag(0, "BogusTag"); ok {
		t.Fatal("expected unknown tag miss")
	}
	if _, ok := TokenByTag(99, "Sync"); ok {
		t.Fatal("expected unknown page miss")
	}
}
